package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// ---------------------------------------------------------------------------
// CONFIG
//
// Nothing is read from or written to local disk anymore. Everything lives in
// three Discord channels, set via env vars:
//
//   DISCORD_BOT_TOKEN
//   DISCORD_MATCHES_CHANNEL_ID  — collector posts one {matchid}.json per match
//   DISCORD_MMR_CHANNEL_ID      — collector posts {puuid}_latest.json snapshots
//   DISCORD_STORE_CHANNEL_ID    — this program's own scratch space: holds the
//                                 current sync_state.json and valorant_details.csv
//                                 as attachments on the two most recent messages
//                                 it posted there. Nothing else should post here.
//
// Bot needs: View Channel + Read Message History on the matches/mmr channels,
// and View Channel + Read Message History + Send Messages + Attach Files +
// Manage Messages (to delete its own superseded state/csv messages) on the
// store channel.
// ---------------------------------------------------------------------------

const discordAPIBase = "https://discord.com/api/v10"
const discordCSVFilename = "valorant_details.csv"
const discordStateFilename = "sync_state.json"

var allyPUUIDs = map[string]bool{
	"9ac37245-e47a-5977-9785-7c2590e2dcda": true,
	"59bae8f3-025c-5dcc-9a1c-c903279e4145": true,
	"5775df6c-0f15-5e6c-8f00-3398dc77d351": true,
	"c38ffa2e-ce9d-5d95-9399-3a89c6af6b16": true,
	"f91099e8-a14b-5913-b66b-13717562a6eb": true,
	"8cc158a4-9733-5b20-90aa-8eac77121911": true,
	"462a8089-15f8-5365-8e52-e0759a870abe": true,
	"663569fe-cd21-52b8-abc0-f9596fc7d1bf": true,
}

var agentRole = map[string]string{
	"Jett": "duelist", "Raze": "duelist", "Reyna": "duelist", "Phoenix": "duelist",
	"Neon": "duelist", "Yoru": "duelist", "Iso": "duelist",
	"Sova": "initiator", "Skye": "initiator", "Breach": "initiator", "Kayo": "initiator",
	"Fade": "initiator", "Gekko": "initiator", "Tejo": "initiator",
	"Omen": "controller", "Brimstone": "controller", "Viper": "controller",
	"Astra": "controller", "Harbor": "controller", "Clove": "controller",
	"Killjoy": "sentinel", "Cypher": "sentinel", "Sage": "sentinel",
	"Chamber": "sentinel", "Deadlock": "sentinel", "Vyse": "sentinel",
}

var healerAgents = map[string]bool{
	"Sage": true, "Skye": true,
}

// ---------------------------------------------------------------------------
// RAW DATA TYPES (unchanged — mirrors what the collector posts as JSON)
// ---------------------------------------------------------------------------

type RawFile struct {
	HasWon    bool `json:"hasWon"`
	MatchData struct {
		Status int       `json:"status"`
		Data   MatchData `json:"data"`
	} `json:"matchData"`
}

type MatchData struct {
	Metadata     Metadata `json:"metadata"`
	Players      Players  `json:"players"`
	Teams        Teams    `json:"teams"`
	AnchorHasWon bool     `json:"-"`
}

type Metadata struct {
	Map              string `json:"map"`
	GameVersion      string `json:"game_version"`
	GameStart        int64  `json:"game_start"`
	GameStartPatched string `json:"game_start_patched"`
	RoundsPlayed     int    `json:"rounds_played"`
	Matchid          string `json:"matchid"`
}

type Players struct {
	AllPlayers []Player `json:"all_players"`
}

type Player struct {
	Puuid       string  `json:"puuid"`
	Team        string  `json:"team"`
	Level       int     `json:"level"`
	Character   string  `json:"character"`
	Currenttier int     `json:"currenttier"`
	PartyID     string  `json:"party_id"`
	Stats       Stats   `json:"stats"`
	Economy     Economy `json:"economy"`
}

type Stats struct {
	Score   int `json:"score"`
	Kills   int `json:"kills"`
	Deaths  int `json:"deaths"`
	Assists int `json:"assists"`
}

type Economy struct {
	Spent        EcoAvg `json:"spent"`
	LoadoutValue EcoAvg `json:"loadout_value"`
}

type EcoAvg struct {
	Average float64 `json:"average"`
}

type Teams struct {
	Red  TeamResult `json:"red"`
	Blue TeamResult `json:"blue"`
}

type TeamResult struct {
	HasWon     bool `json:"has_won"`
	RoundsWon  int  `json:"rounds_won"`
	RoundsLost int  `json:"rounds_lost"`
}

// mmrSnapshotFile mirrors the collector's MMR JSON. The puuid itself comes
// from the attachment filename ({puuid}_latest.json), not this struct.
type mmrSnapshotFile struct {
	Data struct {
		CurrentData struct {
			RankingInTier int `json:"ranking_in_tier"`
		} `json:"current_data"`
	} `json:"data"`
}

// ---------------------------------------------------------------------------
// ROLLING STATE — replaces full-history rescans. Persisted to Discord as
// sync_state.json between runs.
// ---------------------------------------------------------------------------

type RollingState struct {
	Streak         int                `json:"streak"`
	CurrentDay     string             `json:"current_day"`
	GamesToday     int                `json:"games_today"`
	RRByPUUID      map[string]float64 `json:"rr_by_puuid"`
	LastMatchMsgID string             `json:"last_match_msg_id"`
	LastMMRMsgID   string             `json:"last_mmr_msg_id"`
}

func newRollingState() *RollingState {
	return &RollingState{RRByPUUID: map[string]float64{}}
}

// consumeForRow returns the streak/games-played-today values that belong on
// THIS match's row (computed from everything strictly before it), then
// updates the running state so the next match sees this one's result.
func (s *RollingState) consumeForRow(won bool, date string) (streakForRow int, gamesTodayForRow int) {
	if s.CurrentDay != date {
		s.CurrentDay = date
		s.GamesToday = 0
	}
	s.GamesToday++
	gamesTodayForRow = s.GamesToday

	streakForRow = s.Streak
	switch {
	case s.Streak == 0:
		if won {
			s.Streak = 1
		} else {
			s.Streak = -1
		}
	case (s.Streak > 0 && won) || (s.Streak < 0 && !won):
		if won {
			s.Streak++
		} else {
			s.Streak--
		}
	default:
		if won {
			s.Streak = 1
		} else {
			s.Streak = -1
		}
	}
	return
}

// ---------------------------------------------------------------------------
// MAIN
// ---------------------------------------------------------------------------

func Transform() {
	printMemStats("Start transform")

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	token := os.Getenv("DISCORD_BOT_TOKEN")
	matchesChannel := os.Getenv("DISCORD_MATCHES_CHANNEL_ID")
	mmrChannel := os.Getenv("DISCORD_MMR_CHANNEL_ID")
	storeChannel := os.Getenv("DISCORD_STORE_CHANNEL_ID")
	if token == "" || matchesChannel == "" || mmrChannel == "" || storeChannel == "" {
		fmt.Println("DISCORD_BOT_TOKEN / DISCORD_MATCHES_CHANNEL_ID / DISCORD_MMR_CHANNEL_ID / DISCORD_STORE_CHANNEL_ID must all be set")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// --- load state ---
	state := newRollingState()
	stateBytes, stateMsgID, stateFound, err := loadLatestDiscordFile(client, token, storeChannel, discordStateFilename)
	if err != nil {
		fmt.Println("error loading state from discord:", err)
		os.Exit(1)
	}
	if stateFound {
		if err := json.Unmarshal(stateBytes, state); err != nil {
			fmt.Println("error parsing state:", err)
			os.Exit(1)
		}
		if state.RRByPUUID == nil {
			state.RRByPUUID = map[string]float64{}
		}
	}

	// --- load existing CSV (for dedupe + to append new rows to) ---
	csvBytes, csvMsgID, csvFound, err := loadLatestDiscordFile(client, token, storeChannel, discordCSVFilename)
	if err != nil {
		fmt.Println("error loading csv from discord:", err)
		os.Exit(1)
	}
	existingIDs := map[string]bool{}
	if csvFound {
		r := csv.NewReader(bytes.NewReader(csvBytes))
		rows, err := r.ReadAll()
		if err != nil {
			fmt.Println("error parsing existing csv:", err)
			os.Exit(1)
		}
		for i, row := range rows {
			if i == 0 || len(row) == 0 {
				continue
			}
			existingIDs[row[0]] = true
		}
	}

	var outBuf bytes.Buffer
	outBuf.Write(csvBytes)
	if !csvFound {
		w := csv.NewWriter(&outBuf)
		w.Write(csvHeader)
		w.Flush()
	} else if outBuf.Len() > 0 {
		b := outBuf.Bytes()
		if b[len(b)-1] != '\n' {
			outBuf.WriteByte('\n')
		}
	}
	writer := csv.NewWriter(&outBuf)

	// --- sync MMR channel: updates state.RRByPUUID incrementally ---
	mmrCount, newMMRWatermark, err := syncDiscordMMR(client, token, mmrChannel, state.LastMMRMsgID, state.RRByPUUID)
	if err != nil {
		fmt.Println("error syncing mmr channel:", err)
		os.Exit(1)
	}
	state.LastMMRMsgID = newMMRWatermark

	// --- sync matches channel: builds + writes new rows incrementally ---
	written, skipped, newMatchWatermark, err := syncDiscordMatches(client, token, matchesChannel, state.LastMatchMsgID, existingIDs, state, writer)
	if err != nil {
		fmt.Println("error syncing matches channel:", err)
		os.Exit(1)
	}
	state.LastMatchMsgID = newMatchWatermark
	writer.Flush()

	fmt.Printf("discord sync: %d new mmr snapshots, wrote %d new rows, skipped %d duplicates\n", mmrCount, written, skipped)

	// --- persist CSV (only if it changed) and state back to Discord ---
	if written > 0 {
		if _, err := postDiscordMessageWithFile(client, token, storeChannel, discordCSVFilename, outBuf.Bytes()); err != nil {
			fmt.Println("error saving csv to discord:", err)
			os.Exit(1)
		}
		if csvMsgID != "" {
			if err := deleteDiscordMessage(client, token, storeChannel, csvMsgID); err != nil {
				fmt.Println("warning: failed to delete old csv message:", err)
			}
		}
	}

	stateOut, _ := json.Marshal(state)
	if _, err := postDiscordMessageWithFile(client, token, storeChannel, discordStateFilename, stateOut); err != nil {
		fmt.Println("error saving state to discord:", err)
		os.Exit(1)
	}
	if stateMsgID != "" {
		if err := deleteDiscordMessage(client, token, storeChannel, stateMsgID); err != nil {
			fmt.Println("warning: failed to delete old state message:", err)
		}
	}

	printMemStats("End Transform")
}

// ---------------------------------------------------------------------------
// DISCORD SYNC — matches and MMR
// ---------------------------------------------------------------------------

// syncDiscordMMR walks new messages in the MMR channel (oldest to newest) and
// updates rrByPUUID in place. The puuid comes from the "{puuid}_latest.json"
// attachment filename.
func syncDiscordMMR(client *http.Client, token, channelID, after string, rrByPUUID map[string]float64) (count int, watermark string, err error) {
	watermark = after
	for {
		msgs, ferr := fetchDiscordMessagesAfter(client, token, channelID, watermark)
		if ferr != nil {
			return count, watermark, ferr
		}
		if len(msgs) == 0 {
			break
		}
		sortMessagesByIDAscending(msgs)
		for _, msg := range msgs {
			for _, att := range msg.Attachments {
				if !strings.HasSuffix(att.Filename, "_latest.json") {
					continue
				}
				puuid := strings.TrimSuffix(att.Filename, "_latest.json")
				body, derr := downloadDiscordAttachment(client, att.URL)
				if derr != nil {
					return count, watermark, derr
				}
				var snap mmrSnapshotFile
				if uerr := json.Unmarshal(body, &snap); uerr != nil {
					fmt.Println("skipping unparseable mmr snapshot", att.Filename, ":", uerr)
					continue
				}
				rrByPUUID[puuid] = float64(snap.Data.CurrentData.RankingInTier)
				count++
			}
			watermark = msg.ID
		}
		if len(msgs) < 100 {
			break
		}
	}
	return count, watermark, nil
}

// syncDiscordMatches walks new messages in the matches channel (oldest to
// newest — required for rolling stats to stay correct), parses each match
// JSON directly in memory, writes a CSV row immediately, and discards the
// bytes. Assumes the collector posts matches in the order they were played.
func syncDiscordMatches(client *http.Client, token, channelID, after string, existingIDs map[string]bool, state *RollingState, writer *csv.Writer) (written, skipped int, watermark string, err error) {
	watermark = after
	for {
		msgs, ferr := fetchDiscordMessagesAfter(client, token, channelID, watermark)
		if ferr != nil {
			return written, skipped, watermark, ferr
		}
		if len(msgs) == 0 {
			break
		}
		sortMessagesByIDAscending(msgs)
		for _, msg := range msgs {
			for _, att := range msg.Attachments {
				if !strings.HasSuffix(att.Filename, ".json") {
					continue
				}
				body, derr := downloadDiscordAttachment(client, att.URL)
				if derr != nil {
					return written, skipped, watermark, derr
				}
				var raw RawFile
				if uerr := json.Unmarshal(body, &raw); uerr != nil {
					fmt.Println("skipping unparseable match file", att.Filename, ":", uerr)
					continue
				}
				m := raw.MatchData.Data
				m.AnchorHasWon = raw.HasWon

				if existingIDs[m.Metadata.Matchid] {
					skipped++
					continue
				}

				allies, _ := splitTeams(m.Players.AllPlayers)
				won := allyWon(m, allies)
				date, _, _ := parseTimestamp(m.Metadata.GameStart)
				rowStreak, rowGamesToday := state.consumeForRow(won, date)

				row := buildRow(m, rowStreak, rowGamesToday, state.RRByPUUID)
				if werr := writer.Write(row); werr != nil {
					fmt.Println("error writing row:", werr)
					continue
				}
				existingIDs[m.Metadata.Matchid] = true
				written++
			}
			watermark = msg.ID
		}
		if len(msgs) < 100 {
			break
		}
	}
	return written, skipped, watermark, nil
}

// ---------------------------------------------------------------------------
// ROW BUILDING
// ---------------------------------------------------------------------------

var csvHeader = []string{
	"match_id", "date", "time_of_day", "day_of_week", "patch", "is_new_patch",
	"map", "starting_side", "games_played_today",
	"ally_net_winstreak", "ally_avg_rank", "ally_avg_acs", "ally_avg_kd", "ally_avg_eco",
	"ally_num_duelists", "ally_num_initiators", "ally_num_controllers", "ally_num_sentinels", "ally_has_healer",
	"enemy_avg_rank", "enemy_avg_acs", "enemy_avg_kd", "enemy_avg_eco",
	"enemy_num_duelists", "enemy_num_initiators", "enemy_num_controllers", "enemy_num_sentinels", "enemy_has_healer",
	"enemy_smurf_suspected",
	"delta_rank", "delta_acs", "delta_kd", "delta_eco", "won",
	"queue_wait_seconds", "ally_is_premade", "ally_agents", "ally_is_comp_standard",
	"ally_avg_rr", "ally_rank_spread", "tilt_flag",
	"enemy_agents", "enemy_is_premade", "enemy_premade_size", "enemy_avg_rr", "enemy_rank_spread",
	"delta_rr", "rounds_won", "rounds_lost", "round_margin",
}

func buildRow(match MatchData, rowStreak int, rowGamesToday int, rrByPUUID map[string]float64) []string {
	allies, enemies := splitTeams(match.Players.AllPlayers)

	startingSide := "defense"
	if teamRed(match.Players.AllPlayers) {
		startingSide = "attack"
	}

	allyAvgRank := avgTier(allies)
	enemyAvgRank := avgTier(enemies)
	allyAvgACS := avgACS(allies, match.Metadata.RoundsPlayed)
	enemyAvgACS := avgACS(enemies, match.Metadata.RoundsPlayed)
	allyAvgKD := avgKD(allies)
	enemyAvgKD := avgKD(enemies)
	allyAvgEco := avgEco(allies)
	enemyAvgEco := avgEco(enemies)

	ad, ai, ac, as := countRoles(allies)
	ed, ei, ec, es := countRoles(enemies)

	allyHasHealer := hasHealer(allies)
	enemyHasHealer := hasHealer(enemies)

	won := allyWon(match, allies)
	roundsWon, roundsLost := allyRoundResult(match, allies)

	allyPremade, _ := partyInfo(allies)
	enemyPremade, enemyPremadeSize := partyInfo(enemies)

	allyAvgRR, allyRRFound := avgRR(allies, rrByPUUID)
	enemyAvgRR, enemyRRFound := avgRR(enemies, rrByPUUID)

	gameDate, timeOfDay, dayOfWeek := parseTimestamp(match.Metadata.GameStart)

	deltaRR := ""
	if allyRRFound && enemyRRFound {
		deltaRR = f2(allyAvgRR - enemyAvgRR)
	}

	row := []string{
		match.Metadata.Matchid,
		gameDate, timeOfDay, dayOfWeek,
		match.Metadata.GameVersion,
		"", // is_new_patch — fill manually or compare against your own patch-date list
		match.Metadata.Map,
		startingSide,
		itoa(rowGamesToday),

		itoa(rowStreak),
		f2(allyAvgRank), f2(allyAvgACS), f2(allyAvgKD), f2(allyAvgEco),
		itoa(ad), itoa(ai), itoa(ac), itoa(as), boolStr(allyHasHealer),

		f2(enemyAvgRank), f2(enemyAvgACS), f2(enemyAvgKD), f2(enemyAvgEco),
		itoa(ed), itoa(ei), itoa(ec), itoa(es), boolStr(enemyHasHealer),

		"", // enemy_smurf_suspected — manual judgment call

		f2(allyAvgRank - enemyAvgRank), f2(allyAvgACS - enemyAvgACS),
		f2(allyAvgKD - enemyAvgKD), f2(allyAvgEco - enemyAvgEco),
		boolStr(won),

		"",                                           // queue_wait_seconds — not in raw data, log manually
		boolStr(allyPremade), joinAgents(allies), "", // ally_is_comp_standard — compare against your usual comp yourself
		rrOrBlank(allyAvgRR, allyRRFound),
		itoa(rankSpread(allies)),
		"", // tilt_flag — manual, pre-queue judgment call

		joinAgents(enemies), boolStr(enemyPremade), itoa(enemyPremadeSize),
		rrOrBlank(enemyAvgRR, enemyRRFound),
		itoa(rankSpread(enemies)),
		deltaRR,
		itoa(roundsWon), itoa(roundsLost), itoa(roundsWon - roundsLost),
	}
	return row
}

// ---------------------------------------------------------------------------
// PURE HELPERS (unchanged from before — operate on a single match only)
// ---------------------------------------------------------------------------

func splitTeams(players []Player) (allies, enemies []Player) {
	for _, p := range players {
		if allyPUUIDs[p.Puuid] {
			allies = append(allies, p)
		} else {
			enemies = append(enemies, p)
		}
	}
	return
}

func teamRed(players []Player) bool {
	for _, p := range players {
		if allyPUUIDs[p.Puuid] {
			return p.Team == "Red"
		}
	}
	return false
}

func avgTier(players []Player) float64 {
	if len(players) == 0 {
		return 0
	}
	sum := 0
	for _, p := range players {
		sum += p.Currenttier
	}
	return float64(sum) / float64(len(players))
}

func avgACS(players []Player, rounds int) float64 {
	if len(players) == 0 || rounds == 0 {
		return 0
	}
	sum := 0
	for _, p := range players {
		sum += p.Stats.Score
	}
	return float64(sum) / float64(rounds) / float64(len(players))
}

func avgKD(players []Player) float64 {
	if len(players) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range players {
		d := p.Stats.Deaths
		if d == 0 {
			d = 1
		}
		sum += float64(p.Stats.Kills) / float64(d)
	}
	return sum / float64(len(players))
}

func avgEco(players []Player) float64 {
	if len(players) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range players {
		sum += p.Economy.LoadoutValue.Average
	}
	return sum / float64(len(players))
}

// avgRR now reads directly from the incrementally-maintained puuid -> RR map
// (state.RRByPUUID) instead of a per-run filesystem scan of MMR snapshots.
func avgRR(players []Player, rrByPUUID map[string]float64) (float64, bool) {
	if len(players) == 0 {
		return 0, false
	}
	sum, count := 0.0, 0
	for _, p := range players {
		if v, ok := rrByPUUID[p.Puuid]; ok {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	return sum / float64(count), true
}

func rrOrBlank(v float64, found bool) string {
	if !found {
		return ""
	}
	return f2(v)
}

func countRoles(players []Player) (duelists, initiators, controllers, sentinels int) {
	for _, p := range players {
		switch agentRole[p.Character] {
		case "duelist":
			duelists++
		case "initiator":
			initiators++
		case "controller":
			controllers++
		case "sentinel":
			sentinels++
		}
	}
	return
}

func hasHealer(players []Player) bool {
	for _, p := range players {
		if healerAgents[p.Character] {
			return true
		}
	}
	return false
}

func allyWon(match MatchData, allies []Player) bool {
	return match.AnchorHasWon
}

func allyRoundResult(match MatchData, allies []Player) (won, lost int) {
	if len(allies) == 0 {
		return 0, 0
	}
	if allies[0].Team == "Red" {
		return match.Teams.Red.RoundsWon, match.Teams.Red.RoundsLost
	}
	return match.Teams.Blue.RoundsWon, match.Teams.Blue.RoundsLost
}

func partyInfo(players []Player) (isPremade bool, largestGroup int) {
	counts := map[string]int{}
	for _, p := range players {
		counts[p.PartyID]++
	}
	for _, c := range counts {
		if c > largestGroup {
			largestGroup = c
		}
	}
	isPremade = largestGroup == len(players) && len(players) > 0
	return
}

func rankSpread(players []Player) int {
	if len(players) == 0 {
		return 0
	}
	min, max := players[0].Currenttier, players[0].Currenttier
	for _, p := range players {
		if p.Currenttier < min {
			min = p.Currenttier
		}
		if p.Currenttier > max {
			max = p.Currenttier
		}
	}
	return max - min
}

func joinAgents(players []Player) string {
	s := ""
	for i, p := range players {
		if i > 0 {
			s += "|"
		}
		s += p.Character
	}
	return s
}

func parseTimestamp(unixSeconds int64) (date, timeOfDay, dayOfWeek string) {
	if unixSeconds == 0 {
		return "", "", ""
	}
	t := time.Unix(unixSeconds, 0).UTC()
	date = t.Format("2006-01-02")
	dayOfWeek = t.Weekday().String()
	hour := t.Hour()
	switch {
	case hour >= 5 && hour < 12:
		timeOfDay = "morning"
	case hour >= 12 && hour < 17:
		timeOfDay = "afternoon"
	case hour >= 17 && hour < 21:
		timeOfDay = "evening"
	default:
		timeOfDay = "night"
	}
	return
}

func f2(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// ---------------------------------------------------------------------------
// DISCORD HTTP LAYER
// ---------------------------------------------------------------------------

type discordAttachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type discordMessage struct {
	ID          string              `json:"id"`
	Attachments []discordAttachment `json:"attachments"`
}

// fetchDiscordMessagesAfter fetches a single page (up to 100) of messages
// posted after the given snowflake, retrying on rate limits.
func fetchDiscordMessagesAfter(client *http.Client, token, channelID, after string) ([]discordMessage, error) {
	url := fmt.Sprintf("%s/channels/%s/messages?limit=100", discordAPIBase, channelID)
	if after != "" {
		url += "&after=" + after
	}
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bot "+token)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 429 {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("discord API %d: %s", resp.StatusCode, string(b))
		}
		var msgs []discordMessage
		err = json.NewDecoder(resp.Body).Decode(&msgs)
		resp.Body.Close()
		return msgs, err
	}
}

// sortMessagesByIDAscending guarantees chronological processing order
// regardless of what order the Discord API returns a page in — required for
// the rolling-stat math to stay correct.
func sortMessagesByIDAscending(msgs []discordMessage) {
	for i := 1; i < len(msgs); i++ {
		for j := i; j > 0; j-- {
			a, _ := strconv.ParseUint(msgs[j-1].ID, 10, 64)
			b, _ := strconv.ParseUint(msgs[j].ID, 10, 64)
			if a > b {
				msgs[j-1], msgs[j] = msgs[j], msgs[j-1]
			} else {
				break
			}
		}
	}
}

func downloadDiscordAttachment(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d fetching attachment", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// loadLatestDiscordFile scans the most recent messages in a channel for an
// attachment with the exact given filename and returns its bytes plus the
// message ID it lived on (so the caller can delete it once superseded).
func loadLatestDiscordFile(client *http.Client, token, channelID, filename string) ([]byte, string, bool, error) {
	url := fmt.Sprintf("%s/channels/%s/messages?limit=50", discordAPIBase, channelID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", false, err
	}
	req.Header.Set("Authorization", "Bot "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", false, fmt.Errorf("discord API %d: %s", resp.StatusCode, string(b))
	}

	var msgs []discordMessage
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		return nil, "", false, err
	}

	// Discord returns newest-first by default when no after/before is given,
	// so the first match found is the current one.
	for _, msg := range msgs {
		for _, att := range msg.Attachments {
			if att.Filename == filename {
				body, err := downloadDiscordAttachment(client, att.URL)
				if err != nil {
					return nil, "", false, err
				}
				return body, msg.ID, true, nil
			}
		}
	}
	return nil, "", false, nil
}

// postDiscordMessageWithFile uploads data as a new message attachment and
// returns the new message's ID.
func postDiscordMessageWithFile(client *http.Client, token, channelID, filename string, data []byte) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	payloadBytes, _ := json.Marshal(map[string]interface{}{"content": ""})
	pw, err := w.CreateFormField("payload_json")
	if err != nil {
		return "", err
	}
	if _, err := pw.Write(payloadBytes); err != nil {
		return "", err
	}

	fw, err := w.CreateFormFile("files[0]", filename)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(data); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/channels/%s/messages", discordAPIBase, channelID)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("discord post %d: %s", resp.StatusCode, string(body))
	}

	var msg discordMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return "", err
	}
	return msg.ID, nil
}

func deleteDiscordMessage(client *http.Client, token, channelID, messageID string) error {
	url := fmt.Sprintf("%s/channels/%s/messages/%s", discordAPIBase, channelID, messageID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord delete %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

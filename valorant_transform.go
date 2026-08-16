package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// ---------------------------------------------------------------------------
// CONFIG
// ---------------------------------------------------------------------------

const (
	rawMatchesDir = "raw_matches"       // output of the collector script
	rawMMRDir     = "raw_mmr_snapshots" // output of the collector script
	csvPath       = "valorant_details.csv"
)

// Your 5 PUUIDs. Fill these in (same as the collector script's stack, resolved
// to PUUIDs — you can copy them from a raw match file's all_players list, or
// print them out when the collector runs).
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
// RAW DATA TYPES (mirrors what the collector script wrote to disk)
// ---------------------------------------------------------------------------

// RawFile mirrors the actual shape of files in raw_matches/: a top-level
// hasWon flag (from the anchor account's perspective) plus the full API
// response nested under matchData.data.
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
	AnchorHasWon bool     `json:"-"` // populated from the file's top-level hasWon field
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

// MMR snapshot, as written by the collector script (one "latest" file per PUUID)
type MMRSnapshot struct {
	Data struct {
		Name        string `json:"name"`
		Tag         string `json:"tag"`
		CurrentData struct {
			Currenttier   int `json:"currenttier"`
			RankingInTier int `json:"ranking_in_tier"`
		} `json:"current_data"`
	} `json:"data"`
}

// ---------------------------------------------------------------------------
// MAIN
// ---------------------------------------------------------------------------

func Transform() {
	matches, err := loadRawMatches(rawMatchesDir)
	if err != nil {
		fmt.Println("error loading raw matches:", err)
		os.Exit(1)
	}
	if len(matches) == 0 {
		fmt.Println("no raw matches found in", rawMatchesDir)
		os.Exit(1)
	}

	// oldest first, so rolling stats (streak, games_played_today) compute correctly
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Metadata.GameStart < matches[j].Metadata.GameStart
	})

	rrByPUUID := loadRRSnapshots(rawMMRDir)

	existingIDs, err := loadExistingMatchIDs(csvPath)
	if err != nil {
		fmt.Println("error reading existing csv:", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(csvPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("error opening csv:", err)
		os.Exit(1)
	}
	defer f.Close()

	if stat, _ := f.Stat(); stat.Size() == 0 {
		w := csv.NewWriter(f)
		w.Write(csvHeader)
		w.Flush()
	}

	writer := csv.NewWriter(f)
	defer writer.Flush()

	var processed []MatchData // history accumulated as we go, for rolling stats
	written, skipped := 0, 0

	for _, m := range matches {
		if existingIDs[m.Metadata.Matchid] {
			processed = append(processed, m) // still needed for rolling stats continuity
			skipped++
			continue
		}

		row := buildRow(m, processed, rrByPUUID)
		if err := writer.Write(row); err != nil {
			fmt.Println("error writing row:", err)
			continue
		}
		writer.Flush()

		processed = append(processed, m)
		written++
	}

	fmt.Printf("wrote %d new rows, skipped %d already in csv\n", written, skipped)
}

// ---------------------------------------------------------------------------
// LOADING RAW DATA
// ---------------------------------------------------------------------------

func loadRawMatches(dir string) ([]MatchData, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var matches []MatchData
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			fmt.Println("skipping unreadable file", f, ":", err)
			continue
		}
		var raw RawFile
		if err := json.Unmarshal(b, &raw); err != nil {
			fmt.Println("skipping unparseable file", f, ":", err)
			continue
		}
		m := raw.MatchData.Data
		m.AnchorHasWon = raw.HasWon
		matches = append(matches, m)
	}
	return matches, nil
}

// loadRRSnapshots reads the "latest" MMR snapshot per PUUID. Note: this gives
// each player's RANK RIGHT NOW, not their rank at the time of each historical
// match. It's accurate for matches close to when the collector last ran, and
// approximate for older backfilled matches. For true historical RR per match,
// extend the collector to call GetMMRHistoryByPUUID and match by match_id.
func loadRRSnapshots(dir string) map[string]MMRSnapshot {
	result := map[string]MMRSnapshot{}
	files, err := filepath.Glob(filepath.Join(dir, "*_latest.json"))
	if err != nil {
		return result
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var snap MMRSnapshot
		if err := json.Unmarshal(b, &snap); err != nil {
			continue
		}
		puuid := filepath.Base(f)
		puuid = puuid[:len(puuid)-len("_latest.json")]
		result[puuid] = snap
	}
	return result
}

func loadExistingMatchIDs(path string) (map[string]bool, error) {
	ids := map[string]bool{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ids, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	for i, row := range rows {
		if i == 0 || len(row) == 0 {
			continue
		}
		ids[row[0]] = true
	}
	return ids, nil
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

func buildRow(match MatchData, history []MatchData, rrByPUUID map[string]MMRSnapshot) []string {
	allies, enemies := splitTeams(match.Players.AllPlayers)

	team := teamRed(match.Players.AllPlayers)
	startingSide := ""
	if team {
		startingSide = "attack"
	} else {
		startingSide = "defense"
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
		itoa(gamesPlayedToday(history, gameDate)),

		itoa(netWinstreak(history)),
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
// HELPERS
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
			if p.Team == "Red" {
				return true
			} else {
				return false
			}
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

func avgRR(players []Player, rrByPUUID map[string]MMRSnapshot) (float64, bool) {
	if len(players) == 0 {
		return 0, false
	}
	sum, count := 0, 0
	for _, p := range players {
		if snap, ok := rrByPUUID[p.Puuid]; ok {
			sum += snap.Data.CurrentData.RankingInTier
			count++
		}
	}
	if count == 0 {
		return 0, false
	}
	return float64(sum) / float64(count), true
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
	// Prefer the file's top-level hasWon flag — it's given directly and
	// avoids any ambiguity about which color your team was.
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
	t := time.Unix(unixSeconds, 0).UTC() // adjust timezone below if you want local time
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

func gamesPlayedToday(history []MatchData, date string) int {
	count := 1
	for _, m := range history {
		d, _, _ := parseTimestamp(m.Metadata.GameStart)
		if d == date {
			count++
		}
	}
	return count
}

func netWinstreak(history []MatchData) int {
	streak := 0
	for i := len(history) - 1; i >= 0; i-- {
		m := history[i]
		allies, _ := splitTeams(m.Players.AllPlayers)
		won := allyWon(m, allies)
		if streak == 0 {
			if won {
				streak = 1
			} else {
				streak = -1
			}
			continue
		}
		if (streak > 0 && won) || (streak < 0 && !won) {
			if won {
				streak++
			} else {
				streak--
			}
		} else {
			break
		}
	}
	return streak
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

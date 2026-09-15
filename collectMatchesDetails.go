package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"valorant-scrapper/discord"

	"github.com/joho/godotenv"
)

type MatchDetails struct {
	HasWon    bool          `json:"hasWon"`
	MatchData MatchResponse `json:"matchData"`
}

type Rate struct {
	Remaining int
	Used      int
	Reset     int
}

type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details"`
}

type MatchResponse struct {
	Status int `json:"status"`
	Data   struct {
		Metadata struct {
			Map              string `json:"map"`
			GameVersion      string `json:"game_version"`
			GameLength       int    `json:"game_length"`
			GameStart        int    `json:"game_start"`
			GameStartPatched string `json:"game_start_patched"`
			RoundsPlayed     int    `json:"rounds_played"`
			Mode             string `json:"mode"`
			ModeID           string `json:"mode_id"`
			Queue            string `json:"queue"`
			SeasonID         string `json:"season_id"`
			Platform         string `json:"platform"`
			Matchid          string `json:"matchid"`
			PremierInfo      struct {
				TournamentID string `json:"tournament_id"`
				MatchupID    string `json:"matchup_id"`
			} `json:"premier_info"`
			Region  string `json:"region"`
			Cluster string `json:"cluster"`
		} `json:"metadata"`
		Players struct {
			AllPlayers []struct {
				Puuid              string `json:"puuid"`
				Name               string `json:"name"`
				Tag                string `json:"tag"`
				Team               string `json:"team"`
				Level              int    `json:"level"`
				Character          string `json:"character"`
				Currenttier        int    `json:"currenttier"`
				CurrenttierPatched string `json:"currenttier_patched"`
				PlayerCard         string `json:"player_card"`
				PlayerTitle        string `json:"player_title"`
				PartyID            string `json:"party_id"`
				SessionPlaytime    struct {
					Minutes      int `json:"minutes"`
					Seconds      int `json:"seconds"`
					Milliseconds int `json:"milliseconds"`
				} `json:"session_playtime"`
				Assets struct {
					Card struct {
						Small string `json:"small"`
						Large string `json:"large"`
						Wide  string `json:"wide"`
					} `json:"card"`
					Agent struct {
						Small    string `json:"small"`
						Full     string `json:"full"`
						Bust     string `json:"bust"`
						Killfeed string `json:"killfeed"`
					} `json:"agent"`
				} `json:"assets"`
				Behaviour struct {
					AfkRounds    int `json:"afk_rounds"`
					FriendlyFire struct {
						Incoming int `json:"incoming"`
						Outgoing int `json:"outgoing"`
					} `json:"friendly_fire"`
					RoundsInSpawn int `json:"rounds_in_spawn"`
				} `json:"behaviour"`
				Platform struct {
					Type string `json:"type"`
					Os   struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"os"`
				} `json:"platform"`
				AbilityCasts struct {
					CCast int `json:"c_cast"`
					QCast int `json:"q_cast"`
					ECast int `json:"e_cast"`
					XCast int `json:"x_cast"`
				} `json:"ability_casts"`
				Stats struct {
					Score     int `json:"score"`
					Kills     int `json:"kills"`
					Deaths    int `json:"deaths"`
					Assists   int `json:"assists"`
					Bodyshots int `json:"bodyshots"`
					Headshots int `json:"headshots"`
					Legshots  int `json:"legshots"`
				} `json:"stats"`
				Economy struct {
					Spent struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"spent"`
					LoadoutValue struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"loadout_value"`
				} `json:"economy"`
				DamageMade     int `json:"damage_made"`
				DamageReceived int `json:"damage_received"`
			} `json:"all_players"`
			Red []struct {
				Puuid              string `json:"puuid"`
				Name               string `json:"name"`
				Tag                string `json:"tag"`
				Team               string `json:"team"`
				Level              int    `json:"level"`
				Character          string `json:"character"`
				Currenttier        int    `json:"currenttier"`
				CurrenttierPatched string `json:"currenttier_patched"`
				PlayerCard         string `json:"player_card"`
				PlayerTitle        string `json:"player_title"`
				PartyID            string `json:"party_id"`
				SessionPlaytime    struct {
					Minutes      int `json:"minutes"`
					Seconds      int `json:"seconds"`
					Milliseconds int `json:"milliseconds"`
				} `json:"session_playtime"`
				Assets struct {
					Card struct {
						Small string `json:"small"`
						Large string `json:"large"`
						Wide  string `json:"wide"`
					} `json:"card"`
					Agent struct {
						Small    string `json:"small"`
						Full     string `json:"full"`
						Bust     string `json:"bust"`
						Killfeed string `json:"killfeed"`
					} `json:"agent"`
				} `json:"assets"`
				Behaviour struct {
					AfkRounds    int `json:"afk_rounds"`
					FriendlyFire struct {
						Incoming int `json:"incoming"`
						Outgoing int `json:"outgoing"`
					} `json:"friendly_fire"`
					RoundsInSpawn int `json:"rounds_in_spawn"`
				} `json:"behaviour"`
				Platform struct {
					Type string `json:"type"`
					Os   struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"os"`
				} `json:"platform"`
				AbilityCasts struct {
					CCast int `json:"c_cast"`
					QCast int `json:"q_cast"`
					ECast int `json:"e_cast"`
					XCast int `json:"x_cast"`
				} `json:"ability_casts"`
				Stats struct {
					Score     int `json:"score"`
					Kills     int `json:"kills"`
					Deaths    int `json:"deaths"`
					Assists   int `json:"assists"`
					Bodyshots int `json:"bodyshots"`
					Headshots int `json:"headshots"`
					Legshots  int `json:"legshots"`
				} `json:"stats"`
				Economy struct {
					Spent struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"spent"`
					LoadoutValue struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"loadout_value"`
				} `json:"economy"`
				DamageMade     int `json:"damage_made"`
				DamageReceived int `json:"damage_received"`
			} `json:"red"`
			Blue []struct {
				Puuid              string `json:"puuid"`
				Name               string `json:"name"`
				Tag                string `json:"tag"`
				Team               string `json:"team"`
				Level              int    `json:"level"`
				Character          string `json:"character"`
				Currenttier        int    `json:"currenttier"`
				CurrenttierPatched string `json:"currenttier_patched"`
				PlayerCard         string `json:"player_card"`
				PlayerTitle        string `json:"player_title"`
				PartyID            string `json:"party_id"`
				SessionPlaytime    struct {
					Minutes      int `json:"minutes"`
					Seconds      int `json:"seconds"`
					Milliseconds int `json:"milliseconds"`
				} `json:"session_playtime"`
				Assets struct {
					Card struct {
						Small string `json:"small"`
						Large string `json:"large"`
						Wide  string `json:"wide"`
					} `json:"card"`
					Agent struct {
						Small    string `json:"small"`
						Full     string `json:"full"`
						Bust     string `json:"bust"`
						Killfeed string `json:"killfeed"`
					} `json:"agent"`
				} `json:"assets"`
				Behaviour struct {
					AfkRounds    int `json:"afk_rounds"`
					FriendlyFire struct {
						Incoming int `json:"incoming"`
						Outgoing int `json:"outgoing"`
					} `json:"friendly_fire"`
					RoundsInSpawn int `json:"rounds_in_spawn"`
				} `json:"behaviour"`
				Platform struct {
					Type string `json:"type"`
					Os   struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"os"`
				} `json:"platform"`
				AbilityCasts struct {
					CCast int `json:"c_cast"`
					QCast int `json:"q_cast"`
					ECast int `json:"e_cast"`
					XCast int `json:"x_cast"`
				} `json:"ability_casts"`
				Stats struct {
					Score     int `json:"score"`
					Kills     int `json:"kills"`
					Deaths    int `json:"deaths"`
					Assists   int `json:"assists"`
					Bodyshots int `json:"bodyshots"`
					Headshots int `json:"headshots"`
					Legshots  int `json:"legshots"`
				} `json:"stats"`
				Economy struct {
					Spent struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"spent"`
					LoadoutValue struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"loadout_value"`
				} `json:"economy"`
				DamageMade     int `json:"damage_made"`
				DamageReceived int `json:"damage_received"`
			} `json:"blue"`
		} `json:"players"`
		Observers []struct {
			Puuid    string `json:"puuid"`
			Name     string `json:"name"`
			Tag      string `json:"tag"`
			Platform struct {
				Type string `json:"type"`
				Os   struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"os"`
			} `json:"platform"`
			SessionPlaytime struct {
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			} `json:"session_playtime"`
			Team        string `json:"team"`
			Level       int    `json:"level"`
			PlayerCard  string `json:"player_card"`
			PlayerTitle string `json:"player_title"`
			PartyID     string `json:"party_id"`
		} `json:"observers"`
		Coaches []struct {
			Puuid string `json:"puuid"`
			Team  string `json:"team"`
		} `json:"coaches"`
		Teams struct {
			Red struct {
				HasWon     bool `json:"has_won"`
				RoundsWon  int  `json:"rounds_won"`
				RoundsLost int  `json:"rounds_lost"`
				Roaster    struct {
					Members       []string `json:"members"`
					Name          string   `json:"name"`
					Tag           string   `json:"tag"`
					Customization struct {
						Icon      string `json:"icon"`
						Image     string `json:"image"`
						Primary   string `json:"primary"`
						Secondary string `json:"secondary"`
						Tertiary  string `json:"tertiary"`
					} `json:"customization"`
				} `json:"roaster"`
			} `json:"red"`
			Blue struct {
				HasWon     bool `json:"has_won"`
				RoundsWon  int  `json:"rounds_won"`
				RoundsLost int  `json:"rounds_lost"`
				Roaster    struct {
					Members       []string `json:"members"`
					Name          string   `json:"name"`
					Tag           string   `json:"tag"`
					Customization struct {
						Icon      string `json:"icon"`
						Image     string `json:"image"`
						Primary   string `json:"primary"`
						Secondary string `json:"secondary"`
						Tertiary  string `json:"tertiary"`
					} `json:"customization"`
				} `json:"roaster"`
			} `json:"blue"`
		} `json:"teams"`
		Rounds []struct {
			WinningTeam string `json:"winning_team"`
			EndType     string `json:"end_type"`
			BombPlanted bool   `json:"bomb_planted"`
			BombDefused bool   `json:"bomb_defused"`
			PlantEvents struct {
				PlantLocation struct {
					X int `json:"x"`
					Y int `json:"y"`
				} `json:"plant_location"`
				PlantedBy struct {
					Puuid       string `json:"puuid"`
					DisplayName string `json:"display_name"`
					Team        string `json:"team"`
				} `json:"planted_by"`
				PlantSite              string `json:"plant_site"`
				PlantTimeInRound       int    `json:"plant_time_in_round"`
				PlayerLocationsOnPlant []struct {
					PlayerPuuid       string `json:"player_puuid"`
					PlayerDisplayName string `json:"player_display_name"`
					PlayerTeam        string `json:"player_team"`
					Location          struct {
						X int `json:"x"`
						Y int `json:"y"`
					} `json:"location"`
					ViewRadians float64 `json:"view_radians"`
				} `json:"player_locations_on_plant"`
			} `json:"plant_events"`
			DefuseEvents struct {
				DefuseLocation struct {
					X int `json:"x"`
					Y int `json:"y"`
				} `json:"defuse_location"`
				DefusedBy struct {
					Puuid       string `json:"puuid"`
					DisplayName string `json:"display_name"`
					Team        string `json:"team"`
				} `json:"defused_by"`
				DefuseTimeInRound       int `json:"defuse_time_in_round"`
				PlayerLocationsOnDefuse []struct {
					PlayerPuuid       string `json:"player_puuid"`
					PlayerDisplayName string `json:"player_display_name"`
					PlayerTeam        string `json:"player_team"`
					Location          struct {
						X int `json:"x"`
						Y int `json:"y"`
					} `json:"location"`
					ViewRadians float64 `json:"view_radians"`
				} `json:"player_locations_on_defuse"`
			} `json:"defuse_events"`
			PlayerStats []struct {
				AbilityCasts struct {
					CCasts int `json:"c_casts"`
					QCasts int `json:"q_casts"`
					ECasts int `json:"e_casts"`
					XCasts int `json:"x_casts"`
				} `json:"ability_casts"`
				PlayerPuuid       string        `json:"player_puuid"`
				PlayerDisplayName string        `json:"player_display_name"`
				PlayerTeam        string        `json:"player_team"`
				DamageEvents      []interface{} `json:"damage_events"`
				Damage            int           `json:"damage"`
				Bodyshots         int           `json:"bodyshots"`
				Headshots         int           `json:"headshots"`
				Legshots          int           `json:"legshots"`
				KillEvents        []interface{} `json:"kill_events"`
				Kills             int           `json:"kills"`
				Score             int           `json:"score"`
				Economy           struct {
					LoadoutValue int `json:"loadout_value"`
					Weapon       struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Assets struct {
							DisplayIcon  string `json:"display_icon"`
							KillfeedIcon string `json:"killfeed_icon"`
						} `json:"assets"`
					} `json:"weapon"`
					Armor struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Assets struct {
							DisplayIcon string `json:"display_icon"`
						} `json:"assets"`
					} `json:"armor"`
					Remaining int `json:"remaining"`
					Spent     int `json:"spent"`
				} `json:"economy"`
				WasAfk        bool `json:"was_afk"`
				WasPenalized  bool `json:"was_penalized"`
				StayedInSpawn bool `json:"stayed_in_spawn"`
			} `json:"player_stats"`
		} `json:"rounds"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}

type MatchResponses struct {
	Status int `json:"status"`
	Data   []struct {
		Metadata struct {
			Map              string `json:"map"`
			GameVersion      string `json:"game_version"`
			GameLength       int    `json:"game_length"`
			GameStart        int    `json:"game_start"`
			GameStartPatched string `json:"game_start_patched"`
			RoundsPlayed     int    `json:"rounds_played"`
			Mode             string `json:"mode"`
			ModeID           string `json:"mode_id"`
			Queue            string `json:"queue"`
			SeasonID         string `json:"season_id"`
			Platform         string `json:"platform"`
			Matchid          string `json:"matchid"`
			PremierInfo      struct {
				TournamentID string `json:"tournament_id"`
				MatchupID    string `json:"matchup_id"`
			} `json:"premier_info"`
			Region  string `json:"region"`
			Cluster string `json:"cluster"`
		} `json:"metadata"`
		Players struct {
			AllPlayers []struct {
				Puuid              string `json:"puuid"`
				Name               string `json:"name"`
				Tag                string `json:"tag"`
				Team               string `json:"team"`
				Level              int    `json:"level"`
				Character          string `json:"character"`
				Currenttier        int    `json:"currenttier"`
				CurrenttierPatched string `json:"currenttier_patched"`
				PlayerCard         string `json:"player_card"`
				PlayerTitle        string `json:"player_title"`
				PartyID            string `json:"party_id"`
				SessionPlaytime    struct {
					Minutes      int `json:"minutes"`
					Seconds      int `json:"seconds"`
					Milliseconds int `json:"milliseconds"`
				} `json:"session_playtime"`
				Assets struct {
					Card struct {
						Small string `json:"small"`
						Large string `json:"large"`
						Wide  string `json:"wide"`
					} `json:"card"`
					Agent struct {
						Small    string `json:"small"`
						Full     string `json:"full"`
						Bust     string `json:"bust"`
						Killfeed string `json:"killfeed"`
					} `json:"agent"`
				} `json:"assets"`
				Behaviour struct {
					AfkRounds    int `json:"afk_rounds"`
					FriendlyFire struct {
						Incoming int `json:"incoming"`
						Outgoing int `json:"outgoing"`
					} `json:"friendly_fire"`
					RoundsInSpawn int `json:"rounds_in_spawn"`
				} `json:"behaviour"`
				Platform struct {
					Type string `json:"type"`
					Os   struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"os"`
				} `json:"platform"`
				AbilityCasts struct {
					CCast int `json:"c_cast"`
					QCast int `json:"q_cast"`
					ECast int `json:"e_cast"`
					XCast int `json:"x_cast"`
				} `json:"ability_casts"`
				Stats struct {
					Score     int `json:"score"`
					Kills     int `json:"kills"`
					Deaths    int `json:"deaths"`
					Assists   int `json:"assists"`
					Bodyshots int `json:"bodyshots"`
					Headshots int `json:"headshots"`
					Legshots  int `json:"legshots"`
				} `json:"stats"`
				Economy struct {
					Spent struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"spent"`
					LoadoutValue struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"loadout_value"`
				} `json:"economy"`
				DamageMade     int `json:"damage_made"`
				DamageReceived int `json:"damage_received"`
			} `json:"all_players"`
			Red []struct {
				Puuid              string `json:"puuid"`
				Name               string `json:"name"`
				Tag                string `json:"tag"`
				Team               string `json:"team"`
				Level              int    `json:"level"`
				Character          string `json:"character"`
				Currenttier        int    `json:"currenttier"`
				CurrenttierPatched string `json:"currenttier_patched"`
				PlayerCard         string `json:"player_card"`
				PlayerTitle        string `json:"player_title"`
				PartyID            string `json:"party_id"`
				SessionPlaytime    struct {
					Minutes      int `json:"minutes"`
					Seconds      int `json:"seconds"`
					Milliseconds int `json:"milliseconds"`
				} `json:"session_playtime"`
				Assets struct {
					Card struct {
						Small string `json:"small"`
						Large string `json:"large"`
						Wide  string `json:"wide"`
					} `json:"card"`
					Agent struct {
						Small    string `json:"small"`
						Full     string `json:"full"`
						Bust     string `json:"bust"`
						Killfeed string `json:"killfeed"`
					} `json:"agent"`
				} `json:"assets"`
				Behaviour struct {
					AfkRounds    int `json:"afk_rounds"`
					FriendlyFire struct {
						Incoming int `json:"incoming"`
						Outgoing int `json:"outgoing"`
					} `json:"friendly_fire"`
					RoundsInSpawn int `json:"rounds_in_spawn"`
				} `json:"behaviour"`
				Platform struct {
					Type string `json:"type"`
					Os   struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"os"`
				} `json:"platform"`
				AbilityCasts struct {
					CCast int `json:"c_cast"`
					QCast int `json:"q_cast"`
					ECast int `json:"e_cast"`
					XCast int `json:"x_cast"`
				} `json:"ability_casts"`
				Stats struct {
					Score     int `json:"score"`
					Kills     int `json:"kills"`
					Deaths    int `json:"deaths"`
					Assists   int `json:"assists"`
					Bodyshots int `json:"bodyshots"`
					Headshots int `json:"headshots"`
					Legshots  int `json:"legshots"`
				} `json:"stats"`
				Economy struct {
					Spent struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"spent"`
					LoadoutValue struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"loadout_value"`
				} `json:"economy"`
				DamageMade     int `json:"damage_made"`
				DamageReceived int `json:"damage_received"`
			} `json:"red"`
			Blue []struct {
				Puuid              string `json:"puuid"`
				Name               string `json:"name"`
				Tag                string `json:"tag"`
				Team               string `json:"team"`
				Level              int    `json:"level"`
				Character          string `json:"character"`
				Currenttier        int    `json:"currenttier"`
				CurrenttierPatched string `json:"currenttier_patched"`
				PlayerCard         string `json:"player_card"`
				PlayerTitle        string `json:"player_title"`
				PartyID            string `json:"party_id"`
				SessionPlaytime    struct {
					Minutes      int `json:"minutes"`
					Seconds      int `json:"seconds"`
					Milliseconds int `json:"milliseconds"`
				} `json:"session_playtime"`
				Assets struct {
					Card struct {
						Small string `json:"small"`
						Large string `json:"large"`
						Wide  string `json:"wide"`
					} `json:"card"`
					Agent struct {
						Small    string `json:"small"`
						Full     string `json:"full"`
						Bust     string `json:"bust"`
						Killfeed string `json:"killfeed"`
					} `json:"agent"`
				} `json:"assets"`
				Behaviour struct {
					AfkRounds    int `json:"afk_rounds"`
					FriendlyFire struct {
						Incoming int `json:"incoming"`
						Outgoing int `json:"outgoing"`
					} `json:"friendly_fire"`
					RoundsInSpawn int `json:"rounds_in_spawn"`
				} `json:"behaviour"`
				Platform struct {
					Type string `json:"type"`
					Os   struct {
						Name    string `json:"name"`
						Version string `json:"version"`
					} `json:"os"`
				} `json:"platform"`
				AbilityCasts struct {
					CCast int `json:"c_cast"`
					QCast int `json:"q_cast"`
					ECast int `json:"e_cast"`
					XCast int `json:"x_cast"`
				} `json:"ability_casts"`
				Stats struct {
					Score     int `json:"score"`
					Kills     int `json:"kills"`
					Deaths    int `json:"deaths"`
					Assists   int `json:"assists"`
					Bodyshots int `json:"bodyshots"`
					Headshots int `json:"headshots"`
					Legshots  int `json:"legshots"`
				} `json:"stats"`
				Economy struct {
					Spent struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"spent"`
					LoadoutValue struct {
						Overall int     `json:"overall"`
						Average float64 `json:"average"`
					} `json:"loadout_value"`
				} `json:"economy"`
				DamageMade     int `json:"damage_made"`
				DamageReceived int `json:"damage_received"`
			} `json:"blue"`
		} `json:"players"`
		Observers []struct {
			Puuid    string `json:"puuid"`
			Name     string `json:"name"`
			Tag      string `json:"tag"`
			Platform struct {
				Type string `json:"type"`
				Os   struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"os"`
			} `json:"platform"`
			SessionPlaytime struct {
				Minutes      int `json:"minutes"`
				Seconds      int `json:"seconds"`
				Milliseconds int `json:"milliseconds"`
			} `json:"session_playtime"`
			Team        string `json:"team"`
			Level       int    `json:"level"`
			PlayerCard  string `json:"player_card"`
			PlayerTitle string `json:"player_title"`
			PartyID     string `json:"party_id"`
		} `json:"observers"`
		Coaches []struct {
			Puuid string `json:"puuid"`
			Team  string `json:"team"`
		} `json:"coaches"`
		Teams struct {
			Red struct {
				HasWon     bool `json:"has_won"`
				RoundsWon  int  `json:"rounds_won"`
				RoundsLost int  `json:"rounds_lost"`
				Roaster    struct {
					Members       []string `json:"members"`
					Name          string   `json:"name"`
					Tag           string   `json:"tag"`
					Customization struct {
						Icon      string `json:"icon"`
						Image     string `json:"image"`
						Primary   string `json:"primary"`
						Secondary string `json:"secondary"`
						Tertiary  string `json:"tertiary"`
					} `json:"customization"`
				} `json:"roaster"`
			} `json:"red"`
			Blue struct {
				HasWon     bool `json:"has_won"`
				RoundsWon  int  `json:"rounds_won"`
				RoundsLost int  `json:"rounds_lost"`
				Roaster    struct {
					Members       []string `json:"members"`
					Name          string   `json:"name"`
					Tag           string   `json:"tag"`
					Customization struct {
						Icon      string `json:"icon"`
						Image     string `json:"image"`
						Primary   string `json:"primary"`
						Secondary string `json:"secondary"`
						Tertiary  string `json:"tertiary"`
					} `json:"customization"`
				} `json:"roaster"`
			} `json:"blue"`
		} `json:"teams"`
		Rounds []struct {
			WinningTeam string `json:"winning_team"`
			EndType     string `json:"end_type"`
			BombPlanted bool   `json:"bomb_planted"`
			BombDefused bool   `json:"bomb_defused"`
			PlantEvents struct {
				PlantLocation struct {
					X int `json:"x"`
					Y int `json:"y"`
				} `json:"plant_location"`
				PlantedBy struct {
					Puuid       string `json:"puuid"`
					DisplayName string `json:"display_name"`
					Team        string `json:"team"`
				} `json:"planted_by"`
				PlantSite              string `json:"plant_site"`
				PlantTimeInRound       int    `json:"plant_time_in_round"`
				PlayerLocationsOnPlant []struct {
					PlayerPuuid       string `json:"player_puuid"`
					PlayerDisplayName string `json:"player_display_name"`
					PlayerTeam        string `json:"player_team"`
					Location          struct {
						X int `json:"x"`
						Y int `json:"y"`
					} `json:"location"`
					ViewRadians float64 `json:"view_radians"`
				} `json:"player_locations_on_plant"`
			} `json:"plant_events"`
			DefuseEvents struct {
				DefuseLocation struct {
					X int `json:"x"`
					Y int `json:"y"`
				} `json:"defuse_location"`
				DefusedBy struct {
					Puuid       string `json:"puuid"`
					DisplayName string `json:"display_name"`
					Team        string `json:"team"`
				} `json:"defused_by"`
				DefuseTimeInRound       int `json:"defuse_time_in_round"`
				PlayerLocationsOnDefuse []struct {
					PlayerPuuid       string `json:"player_puuid"`
					PlayerDisplayName string `json:"player_display_name"`
					PlayerTeam        string `json:"player_team"`
					Location          struct {
						X int `json:"x"`
						Y int `json:"y"`
					} `json:"location"`
					ViewRadians float64 `json:"view_radians"`
				} `json:"player_locations_on_defuse"`
			} `json:"defuse_events"`
			PlayerStats []struct {
				AbilityCasts struct {
					CCasts int `json:"c_casts"`
					QCasts int `json:"q_casts"`
					ECasts int `json:"e_casts"`
					XCasts int `json:"x_casts"`
				} `json:"ability_casts"`
				PlayerPuuid       string        `json:"player_puuid"`
				PlayerDisplayName string        `json:"player_display_name"`
				PlayerTeam        string        `json:"player_team"`
				DamageEvents      []interface{} `json:"damage_events"`
				Damage            int           `json:"damage"`
				Bodyshots         int           `json:"bodyshots"`
				Headshots         int           `json:"headshots"`
				Legshots          int           `json:"legshots"`
				KillEvents        []interface{} `json:"kill_events"`
				Kills             int           `json:"kills"`
				Score             int           `json:"score"`
				Economy           struct {
					LoadoutValue int `json:"loadout_value"`
					Weapon       struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Assets struct {
							DisplayIcon  string `json:"display_icon"`
							KillfeedIcon string `json:"killfeed_icon"`
						} `json:"assets"`
					} `json:"weapon"`
					Armor struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Assets struct {
							DisplayIcon string `json:"display_icon"`
						} `json:"assets"`
					} `json:"armor"`
					Remaining int `json:"remaining"`
					Spent     int `json:"spent"`
				} `json:"economy"`
				WasAfk        bool `json:"was_afk"`
				WasPenalized  bool `json:"was_penalized"`
				StayedInSpawn bool `json:"stayed_in_spawn"`
			} `json:"player_stats"`
		} `json:"rounds"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}

func appendProcessedID(filename, matchID, botToken, channelID string) {
	msg, err := discord.FetchLatestMessage(&http.Client{}, botToken, channelID)
	if err != nil {
		log.Printf("Failed to fetch %s from Discord: %v", filename, err)
	}
	var file []byte
	for _, data := range msg.Attachments {
		file, err = discord.DownloadAttachment(&http.Client{}, data.URL)
		if err != nil {
			log.Printf("Failed to download %s from Discord: %v", filename, err)
		}
	}
	formattedID := matchID + "\n"
	file = append(file, []byte(formattedID)...)
	if len(msg.Attachments) != 0 {
		err = discord.DeleteMessage(&http.Client{}, botToken, channelID, msg.ID)
		if err != nil {
			log.Printf("Failed to delete file %s with id %s beacuse of %v\n", filename, msg.ID, err)
		}
	}

	id, err := discord.PostMessageWithAttachment(&http.Client{}, botToken, channelID, filename, file, "")
	if err != nil {
		log.Printf("Failed to upload file to discord beacuse of %v", err)
	}

	log.Printf("File uploaded to discord with id %s", id)
}

func loadProcessedIDs(filename, botToken, channelID string) map[string]bool {
	processed := make(map[string]bool)
	msg, err := discord.FetchLatestMessage(&http.Client{}, botToken, channelID)
	if err != nil {
		log.Printf("Failed to fetch %s from Discord: %v", filename, err)
		return processed
	}
	var file []byte
	for _, data := range msg.Attachments {
		file, err = discord.DownloadAttachment(&http.Client{}, data.URL)
		if err != nil {
			log.Printf("Failed to download %s from Discord: %v", filename, err)
			return processed
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		id := strings.TrimSpace(scanner.Text())
		if id != "" {
			processed[id] = true
		}
	}
	return processed
}

func CollectMatchesData() {
	printMemStats("Start Collection")
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	apiKey := os.Getenv("API_KEY")
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	channelID := os.Getenv("DISCORD_MATCHES_CHANNEL_ID")
	processedMatchesID := os.Getenv("DISCORD_PROCESSED_MATCHES_CHANNEL_ID")

	if botToken == "" || channelID == "" {
		log.Fatal("Missing Discord credentials in .env")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	rate := &Rate{}

	// 1. Load previously processed Match IDs to prevent duplicate uploads
	processedIDs := loadProcessedIDs("processed_matches.txt", botToken, processedMatchesID)

	url := fmt.Sprintf("https://api.henrikdev.xyz/valorant/v3/by-puuid/matches/%v/%v%v", "ap", "59bae8f3-025c-5dcc-9a1c-c903279e4145", "?mode=competitive&size=10")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	req.Header.Add("Authorization", apiKey)

	if rate.Remaining < 4 && rate.Remaining != 0 { // Added != 0 so it doesn't sleep on first run
		log.Printf("Sleeping for %d second", rate.Reset)
		time.Sleep(time.Duration(rate.Reset) * time.Second)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	rate.Used, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-limit"))
	rate.Remaining, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-remaining"))
	rate.Reset, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-reset"))

	var matchRes MatchResponses
	decodeErr := json.NewDecoder(resp.Body).Decode(&matchRes)
	if decodeErr != nil {
		log.Printf("JSON decode error: %v", decodeErr)
		return
	}

	skipped := 0

	for _, data := range matchRes.Data {
		matchID := data.Metadata.Matchid

		// Check if we already uploaded this match to Discord
		if processedIDs[matchID] {
			skipped++
			continue
		}

		if matchRes.Errors != nil {
			log.Printf("API Errors: %v", matchRes.Errors)
			continue
		}

		var matchDetails MatchDetails
		matchDetails.MatchData.Data = data
		matchDetails.MatchData.Errors = matchRes.Errors
		matchDetails.MatchData.Status = matchRes.Status

		set := false
		for _, player := range data.Players.Blue {
			if player.Puuid == "462a8089-15f8-5365-8e52-e0759a870abe" || player.Puuid == "59bae8f3-025c-5dcc-9a1c-c903279e4145" || player.Puuid == "9ac37245-e47a-5977-9785-7c2590e2dcda" {
				matchDetails.HasWon = data.Teams.Blue.HasWon
				set = true
				break // slightly optimized
			}
		}
		if !set {
			matchDetails.HasWon = data.Teams.Red.HasWon
		}

		outputJSON, err := json.MarshalIndent(matchDetails, "", "\t")
		if err != nil {
			log.Printf("Failed to marshal struct to JSON: %v", err)
			continue
		}

		filename := matchID + ".json"

		// 2. Upload to Discord
		id, err := discord.PostMessageWithAttachment(&http.Client{}, botToken, channelID, filename, outputJSON, "")
		if err != nil {
			log.Printf("Failed to upload %s to Discord: %v", filename, err)
			continue
		}

		log.Printf("Successfully uploaded %s to Discord with ID %s\n", filename, id)

		// 3. Mark as processed locally
		appendProcessedID("processed_matches.txt", matchID, botToken, processedMatchesID)
		processedIDs[matchID] = true // update local map for current run
	}

	log.Printf("Skipped %d already processed matches\n", skipped)
	printMemStats("End Collection")
}

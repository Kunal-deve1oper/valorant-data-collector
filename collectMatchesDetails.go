package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

func CollectMatchesData() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("API_KEY")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	rate := &Rate{}

	entries, err := os.ReadDir("match_id")
	if err != nil {
		panic(err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	url := fmt.Sprintf("https://api.henrikdev.xyz/valorant/v3/by-puuid/matches/%v/%v%v", "ap", "462a8089-15f8-5365-8e52-e0759a870abe", "?mode=competitive&size=10")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}

	req.Header.Add("Authorization", apiKey)

	if rate.Remaining < 4 {
		log.Printf("Sleeping for %d second", rate.Reset)
		time.Sleep(time.Duration(rate.Reset) * time.Second)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
	}

	rate.Used, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-limit"))
	rate.Remaining, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-remaining"))
	rate.Reset, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-reset"))

	var matchRes MatchResponses
	decodeErr := json.NewDecoder(resp.Body).Decode(&matchRes)

	if decodeErr != nil {
		log.Printf("JSON decode error: %v", decodeErr)
	}

	resp.Body.Close()

	skipped := 0

	for _, data := range matchRes.Data {
		outPath := filepath.Join("raw_matches", data.Metadata.Matchid+".json")

		if _, err := os.Stat(outPath); err == nil {
			skipped++
			continue
		}

		if matchRes.Errors != nil {
			log.Printf("%v:%v", err, matchRes.Errors)
			continue
		}

		var matchDetails MatchDetails

		matchDetails.MatchData.Data = data
		matchDetails.MatchData.Errors = matchRes.Errors
		matchDetails.MatchData.Status = matchRes.Status
		set := false
		for _, player := range data.Players.Blue {
			if player.Puuid == "462a8089-15f8-5365-8e52-e0759a870abe" {
				matchDetails.HasWon = data.Teams.Blue.HasWon
				set = true
			}
		}
		if !set {
			matchDetails.HasWon = data.Teams.Red.HasWon
		}

		outputJSON, err := json.MarshalIndent(matchDetails, "", "\t")
		if err != nil {
			log.Fatalf("Failed to marshal struct to JSON: %v", err)
		}

		err = os.WriteFile(outPath, outputJSON, 0644)
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}

		log.Printf("Successfully saved data to %s\n", outPath)

		log.Printf("%v\n", rate)

	}
	log.Printf("Skipped %d\n", skipped)
}

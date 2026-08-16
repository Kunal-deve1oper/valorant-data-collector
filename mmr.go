package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	govapi "github.com/yldshv/go-valorant-api"
)

type PlayerDetails struct {
	Name  string
	Tag   string
	Puuid string
}

func Mmr() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("API_KEY")
	vapi := govapi.New(govapi.WithKey(apiKey))

	playerList := []PlayerDetails{
		{Name: "Ned", Tag: "2933", Puuid: "663569fe-cd21-52b8-abc0-f9596fc7d1bf"},
		{Name: "z0rokillsnoobs", Tag: "2003", Puuid: "8cc158a4-9733-5b20-90aa-8eac77121911"},
		{Name: "SC4R", Tag: "LORD", Puuid: "9ac37245-e47a-5977-9785-7c2590e2dcda"},
		{Name: "systemctl start", Tag: "4575", Puuid: "59bae8f3-025c-5dcc-9a1c-c903279e4145"},
		{Name: "Krypto", Tag: "4284", Puuid: "462a8089-15f8-5365-8e52-e0759a870abe"},
		{Name: "GaramheGaramhe", Tag: "ahhh", Puuid: "5775df6c-0f15-5e6c-8f00-3398dc77d351"},
		{Name: "NoSheat", Tag: "6917", Puuid: "f91099e8-a14b-5913-b66b-13717562a6eb"},
		{Name: "Protein Chut", Tag: "8149", Puuid: "c38ffa2e-ce9d-5d95-9399-3a89c6af6b16"},
	}

	for _, item := range playerList {

		if vapi.Ratelimits.Remaining < 3 {
			log.Printf("Sleeping for %d seconds", vapi.Ratelimits.Reset)
			time.Sleep(time.Duration(vapi.Ratelimits.Reset) * time.Second)
		}

		if item.Puuid == "" {
			account, err := vapi.GetAccountByName(govapi.GetAccountByNameParams{
				Name: "Ned",
				Tag:  "2933",
			})
			item.Puuid = account.Data.Puuid
			if err != nil {
				fmt.Println("error fetching mmr for")
			}
		}

		mmr, err := vapi.GetMMRByPUUIDv2(govapi.GetMMRByPUUIDv2Params{
			Affinity: "ap",
			Puuid:    item.Puuid,
		})
		if err != nil {
			fmt.Println("error fetching mmr for")
		}

		raw, err := json.MarshalIndent(mmr, "", "  ")
		if err != nil {
			fmt.Println("error marshaling mmr for")
		}

		outPath := filepath.Join("raw_mmr_snapshots", item.Puuid+"_latest.json")
		if err := os.WriteFile(outPath, raw, 0644); err != nil {
			fmt.Println("error writing mmr snapshot for", ":", err)
		}
	}
}

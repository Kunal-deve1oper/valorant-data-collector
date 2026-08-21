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

type Matches struct {
	Data Data `json:"data"`
}

func CollectData() {
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

	for i := range count {
		file := filepath.Join("match_id", "matches_next_"+strconv.Itoa(i)+".json")

		f, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Error occured while opening file %s: %v", file, err)
		}

		var matches Matches

		err = json.Unmarshal(f, &matches)
		if err != nil {
			log.Fatalf("Failed to unmarshal json %v", err)
		}

		skipped := 0

		for _, data := range matches.Data.Matches {
			outPath := filepath.Join("raw_matches", data.Attributes.MatchId+".json")

			if _, err := os.Stat(outPath); err == nil {
				skipped++
				continue
			}

			url := fmt.Sprintf("https://api.henrikdev.xyz/valorant/v2/match/%v", data.Attributes.MatchId)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				log.Printf("Error creating request: %v: %s", err, data.Attributes.MatchId)
				continue
			}

			req.Header.Add("Authorization", apiKey)

			if rate.Remaining < 4 {
				log.Printf("Sleeping for %d second", rate.Reset)
				time.Sleep(time.Duration(rate.Reset) * time.Second)
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("HTTP request failed: %v: %s", err, data.Attributes.MatchId)
				continue
			}

			rate.Used, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-limit"))
			rate.Remaining, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-remaining"))
			rate.Reset, _ = strconv.Atoi(resp.Header.Get("x-ratelimit-reset"))

			var matchRes MatchResponse
			decodeErr := json.NewDecoder(resp.Body).Decode(&matchRes)

			resp.Body.Close()

			if decodeErr != nil {
				log.Printf("JSON decode error: %v: %s", decodeErr, data.Attributes.MatchId)
				continue
			}

			if matchRes.Errors != nil {
				log.Printf("%v:%v:%s", err, matchRes.Errors, data.Attributes.MatchId)
				continue
			}

			var matchDetails MatchDetails

			matchDetails.MatchData = matchRes
			if data.Metadata.Result == "victory" {
				matchDetails.HasWon = true
			} else {
				matchDetails.HasWon = false
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

}

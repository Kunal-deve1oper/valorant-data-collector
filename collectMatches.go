package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type APIResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	Matches []Match `json:"matches"`
}

type Match struct {
	Attributes struct {
		MatchId string `json:"id"`
	} `json:"attributes"`
	Metadata struct {
		Timestamp string `json:"timestamp"`
		Result    string `json:"result"`
		MapName   string `josn:"mapName"`
	} `json:"metadata"`
}

func CollectMatches() {
	nextValue := 0

	url := fmt.Sprintf("https://api.tracker.gg/api/v2/valorant/standard/matches/riot/Krypto%%234284?platform=pc&type=competitive&next=%d", nextValue)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Received non-success status code: %d, %v", resp.StatusCode, resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	log.Printf("Successfully parsed %d matches.\n", len(apiResponse.Data.Matches))

	outputJSON, err := json.MarshalIndent(apiResponse, "", "\t")
	if err != nil {
		log.Fatalf("Failed to marshal struct to JSON: %v", err)
	}

	dirName := "match_id"

	err = os.MkdirAll(dirName, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	fileName := fmt.Sprintf("matches_next_%d.json", nextValue)
	fullPath := filepath.Join(dirName, fileName)

	err = os.WriteFile(fullPath, outputJSON, 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	log.Printf("Successfully saved data to %s\n", fileName)
}

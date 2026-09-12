package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

// JobStatus reports the current/last state of the pipeline job.
type JobStatus struct {
	Running    bool      `json:"running"`
	LastRunAt  time.Time `json:"last_run_at"`
	LastRunDur string    `json:"last_run_duration,omitempty"`
	LastError  string    `json:"last_error,omitempty"`
	NextRunAt  time.Time `json:"next_run_at"`
	RunCount   int       `json:"run_count"`
}

var (
	statusMu sync.Mutex
	status   JobStatus
)

func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf(
		"[%s] Alloc=%d MB HeapAlloc=%d MB HeapSys=%d MB Sys=%d MB Objects=%d GC=%d",
		label,
		m.Alloc/1024/1024,
		m.HeapAlloc/1024/1024,
		m.HeapSys/1024/1024,
		m.Sys/1024/1024,
		m.HeapObjects,
		m.NumGC,
	)
}

// runJob executes the full data pipeline (CollectMatchesData -> Mmr -> Transform).
// If a run is already in progress, the new trigger is skipped rather than overlapping.
func runJob() {
	statusMu.Lock()
	if status.Running {
		statusMu.Unlock()
		log.Println("Job already running, skipping this trigger")
		return
	}
	status.Running = true
	statusMu.Unlock()

	start := time.Now()
	printMemStats("Start job")

	var runErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				runErr = fmt.Errorf("panic: %v", r)
			}
		}()

		log.Println("Started collecting match details...")
		CollectMatchesData()
		log.Println("Done collecting match details!!!")

		log.Println("Started collecting mmr details...")
		Mmr()
		log.Println("Done collecting mmr details!!!")

		log.Println("Started transforming data...")
		Transform()
		log.Println("Done transforming data!!!")
	}()

	printMemStats("End job")
	dur := time.Since(start)

	statusMu.Lock()
	status.Running = false
	status.LastRunAt = start
	status.LastRunDur = dur.String()
	status.RunCount++
	if runErr != nil {
		status.LastError = runErr.Error()
		log.Printf("Job failed: %v", runErr)
	} else {
		status.LastError = ""
		log.Printf("Job completed successfully in %s", dur)
	}
	statusMu.Unlock()
}

// durationUntilNext1AM returns how long to wait until the next 1:00 AM local time.
func durationUntilNext1AM() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}

// scheduler runs forever, triggering runJob every day at 1:00 AM local time.
func scheduler() {
	for {
		wait := durationUntilNext1AM()

		statusMu.Lock()
		status.NextRunAt = time.Now().Add(wait)
		statusMu.Unlock()

		log.Printf("Next scheduled run at %s (in %s)", status.NextRunAt.Format(time.RFC3339), wait)
		time.Sleep(wait)
		runJob()
	}
}

func triggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statusMu.Lock()
	running := status.Running
	statusMu.Unlock()

	if running {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"message": "job already running"})
		return
	}

	go runJob()
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "job triggered"})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	statusMu.Lock()
	defer statusMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	printMemStats("Start main")

	// Background scheduler: runs the pipeline automatically every day at 1 AM,
	// as long as the service happens to be awake. On Render's free tier the
	// service spins down after 15 min idle, so this alone isn't reliable —
	// pair it with an external cron ping (see README) hitting /trigger at 1am.
	go scheduler()

	mux := http.NewServeMux()
	mux.HandleFunc("/trigger", triggerHandler) // POST: run the pipeline now
	mux.HandleFunc("/status", statusHandler)   // GET: current/last run info
	mux.HandleFunc("/health", healthHandler)   // GET: liveness check

	// Render injects the PORT env var; the service must bind to it.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Starting API server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

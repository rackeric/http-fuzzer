package main

import (
	"context"
	"net/http"

	"fuzzer/internal/api"
	"fuzzer/internal/fuzzer"
	"fuzzer/internal/logging"
	"fuzzer/internal/storage"
	"fuzzer/internal/wordlist"
)

func main() {
	// Initialize logging (you can pass LevelDebug for more verbose logging)
	logging.InitLogger(logging.LevelInfo)

	ctx := context.Background()

	store, err := storage.NewJobStore("jobs.json")
	if err != nil {
		logging.Error("Failed to initialize job store: %v", err)
		return
	}
	logging.Info("Job store initialized successfully")

	wordlistMgr, err := wordlist.NewManager("wordlists")
	if err != nil {
		logging.Error("Failed to initialize wordlist manager: %v", err)
		return
	}
	logging.Info("Wordlist manager initialized successfully")

	manager := fuzzer.NewManager(ctx, store, wordlistMgr, 10.0)
	logging.Info("Fuzzer manager initialized")

	apiHandler := api.NewHandler(manager, wordlistMgr, store)

	// Serve static files for UI
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/", fs)
	logging.Info("Static file server initialized")

	// API routes
	http.Handle("/api/", apiHandler)
	logging.Info("API routes registered")

	logging.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logging.Error("Server failed to start: %v", err)
	}
}

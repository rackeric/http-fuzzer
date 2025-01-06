package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"fuzzer/api"
	"fuzzer/fuzzer"
	"fuzzer/storage"
	"fuzzer/wordlist"
)

func main() {
	ctx := context.Background()

	store, err := storage.NewJobStore("jobs.json")
	if err != nil {
		fmt.Println("Failed to initialize job store:", err)
		return
	}

	wordlistMgr, err := wordlist.NewManager("wordlists")
	if err != nil {
		fmt.Println("Failed to initialize wordlist manager:", err)
		return
	}

	manager := fuzzer.NewManager(ctx, store, wordlistMgr, 10.0)
	fmt.Println("Fuzzer manager initialized:", manager)

	apiHandler := api.NewHandler(manager, wordlistMgr, store)

	// Serve static files for UI
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API routes
	http.Handle("/api/", apiHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

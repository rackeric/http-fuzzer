package types

import (
	"time"
)

// Define custom types for job types
type JobType string

const (
	DirectoryType JobType = "directory"
	SubdomainType JobType = "subdomain"
)

type Job struct {
	ID         string    `json:"id"`
	Target     string    `json:"target"`
	Type       JobType   `json:"type"`
	WordlistID string    `json:"wordlistId"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	Findings   []Finding `json:"findings"`
	StartTime  time.Time `json:"startTime"`
}

type Finding struct {
	URL   string    `json:"url"`
	Type  string    `json:"type"`
	Found time.Time `json:"found"`
}

type Wordlist struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Words []string `json:"words"`
}

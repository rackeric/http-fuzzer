package types

import (
	"time"
)

// type Job struct {
// 	ID         string    `json:"id"`
// 	Target     string    `json:"target"`
// 	Type       string    // "subdomain", "directory"
// 	WordlistID string    `json:"wordlistId"`
// 	Status     string    `json:"status"` // "running", "paused", "stopped", "completed"
// 	Progress   float64   `json:"progress"`
// 	Findings   []Finding `json:"findings"`
// 	LastUpdate time.Time
// 	CancelFunc context.CancelFunc `json:"cancelFunc"`
// 	PauseChan  chan bool          `json:"pauseChan"`
// }

type Job struct {
	ID         string    `json:"id"`
	Target     string    `json:"target"`
	Type       string    `json:"type"`
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

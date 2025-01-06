package types

type FuzzerManager interface {
	StartJob(target, wordlistID, jobType string) error
	StopJob(jobID string) error
	GetJobs() ([]*Job, error)
}

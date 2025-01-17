package types

type FuzzerManager interface {
	StartJob(target, wordlistID string, jobType JobType) error
	StopJob(jobID string) error
	GetJobs() ([]*Job, error)
}

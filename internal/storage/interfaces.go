package storage

import "fuzzer/types"

type JobStorer interface {
	SaveJob(job *types.Job) error
	GetJob(id string) (*types.Job, error)
	ListJobs() ([]*types.Job, error)
	Save() error
}

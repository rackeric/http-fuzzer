package storage

import "fuzzer/types"

type JobStorer interface {
	DeleteJob(id string) error
	SaveJob(job *types.Job) error
	GetJob(id string) (*types.Job, error)
	ListJobs() ([]*types.Job, error)
	Save() error
}

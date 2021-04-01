package main

import (
	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/structs"
)

type PipelineStatus string

const (
	PipelineStatusIdle    PipelineStatus = "idle"
	PipelineStatusWorking                = "working"
)

type PipelineWorker struct {
	ID         string
	CommitHash string
	Status     PipelineStatus
	Pipeline   *structs.Pipeline
	Agent      *Agent
}

func NewPipelineWorker(p *structs.Pipeline) *PipelineWorker {
	uid, _ := uuid.GenerateUUID()
	w := &PipelineWorker{
		ID:       uid,
		Pipeline: p,
	}

	return w
}

func (w *PipelineWorker) Run() {
	w.Status = PipelineStatusWorking

	// TODO:
	// 1. create pipeline directory
	// 2. archive whole repo
	// 3. expose archive via api
	// 4. iterate through jobs
	// 5. for each job ask for free agent
	// 6. if job return any artifacts, save them to the pipeline dir via exposed api
	// 7. jobs and inner tasks push logs via rtm channel
	// 8. on finish pipeline (or any error) all resources should be cleaned up (like pipeline directory)
}

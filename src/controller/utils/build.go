package utils

import (
	"log"

	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/structs"
)

type PipelineStatus string

const (
	PipelineStatusIdle    PipelineStatus = "idle"
	PipelineStatusWorking                = "working"
)

type PipelineBuild struct {
	ID         string
	CommitHash string
	Status     PipelineStatus
	Pipeline   *structs.Pipeline
	Agent      *Agent

	LogChan    chan string
	ErrLogChan chan string
}

func NewPipelineBuild(p *structs.Pipeline) *PipelineBuild {
	logCh := make(chan string)
	errLogCh := make(chan string)

	uid, _ := uuid.GenerateUUID()
	w := &PipelineBuild{
		ID:         uid,
		Pipeline:   p,
		LogChan:    logCh,
		ErrLogChan: errLogCh,
	}

	go w.logs()

	return w
}

func (w *PipelineBuild) Run() {
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

func (b *PipelineBuild) logs() {
	for {
		select {
		case logStr := <-b.LogChan:
			log.Printf("\033[97m%s\033[0m", logStr)
		case errLogStr := <-b.ErrLogChan:
			log.Printf("\033[31m%s\033[0m", errLogStr)
		}

	}
}

package utils

import (
	"log"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/events/controller"
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
	Triggers   *PipelineTriggers

	LogChan           chan string
	ErrLogChan        chan string
	AgentResponseChan chan *Agent
	askAgentChan      chan string
}

func NewPipelineBuild(p *structs.Pipeline, askAgentCh chan string) *PipelineBuild {
	logCh := make(chan string)
	errLogCh := make(chan string)

	uid, _ := uuid.GenerateUUID()
	w := &PipelineBuild{
		ID:                uid,
		Pipeline:          p,
		LogChan:           logCh,
		ErrLogChan:        errLogCh,
		AgentResponseChan: make(chan *Agent),
		askAgentChan:      askAgentCh,
		Triggers:          NewPipelineTriggers(),
	}

	go w.logs()

	return w
}

func (w *PipelineBuild) Run() {
	w.Status = PipelineStatusWorking
	w.Triggers.RegisterJobsFor(w.Pipeline)

	// TODO:
	// 1. create pipeline directory
	// 2. archive whole repo
	// 3. expose archive via api
	// 4. iterate through jobs
	// 5. for each job ask for free agent
	// 6. if job return any artifacts, save them to the pipeline dir via exposed api
	// 7. jobs and inner tasks push logs via rtm channel
	// 8. on finish pipeline (or any error) all resources should be cleaned up (like pipeline directory)

	for _, j := range w.Pipeline.Jobs {
		// Job with conditions should run in special way
		if len(j.Conditons) != 0 {
			continue
		}

		w.StartJob(&j, false)
		// Another jobs will execute in JobFinishedEventHandler
		return
	}
}

func (b *PipelineBuild) StartJob(job *structs.Job, isConditional bool) {
	b.askAgentChan <- b.ID
	agent := <-b.AgentResponseChan
	b.Agent = agent

	oid, _ := uuid.GenerateUUID()
	ev := &controller.StartJobCommand{
		EventData:     &goeh.EventData{ID: oid},
		BuildID:       b.ID,
		PipelineID:    b.Pipeline.Name,
		Job:           *job,
		IsConditional: isConditional,
	}

	agent.SendMessage(ev)
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

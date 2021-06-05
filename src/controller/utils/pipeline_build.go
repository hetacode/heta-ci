package utils

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/commons"
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

	RepositoryArchivePath string
	ArtifactsDir          string

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
		ID:                    uid,
		Pipeline:              p,
		LogChan:               logCh,
		ErrLogChan:            errLogCh,
		AgentResponseChan:     make(chan *Agent),
		askAgentChan:          askAgentCh,
		Triggers:              NewPipelineTriggers(),
		RepositoryArchivePath: path.Join(RepositoryDirectory, p.RepositoryArchiveID+".zip"),
	}

	go w.logs()

	return w
}

func (w *PipelineBuild) Run() {
	w.Status = PipelineStatusWorking
	w.Triggers.RegisterJobsFor(w.Pipeline)

	// TODO:
	// 8. on finish pipeline (or any error) all resources should be cleaned up (like pipeline directory)

	if err := w.initBuildDirs(w.ID); err != nil {
		w.ErrLogChan <- err.Error()
		return
	}

	for _, j := range w.Pipeline.Jobs {
		// Job with conditions should run in special way
		if len(j.Conditons) != 0 {
			continue
		}

		if err := w.StartJob(&j, false); err != nil {
			w.ErrLogChan <- err.Error()
		}
		// Another jobs will execute in JobFinishedEventHandler
		return
	}
}

func (b *PipelineBuild) StartJob(job *structs.Job, isConditional bool) error {
	artifactsFilePath := b.ArtifactsDir + "/artifacts.zip"
	hasArtifacts, err := commons.IsFileExists(artifactsFilePath)
	if err != nil {
		return fmt.Errorf("get artifacts file exists failed: %s", err)
	}

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
		HasArtifacts:  hasArtifacts,
	}

	agent.SendMessage(ev)

	return nil
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

func (b *PipelineBuild) initBuildDirs(pipelineID string) error {
	pipelineDir := path.Join(PipelinesDir, pipelineID)
	artifactsDir := path.Join(pipelineDir, ArtifactsDir)

	if err := os.Mkdir(pipelineDir, 0777); err != nil {
		return fmt.Errorf("start pipeline build | create %s dir failed | err: %s", pipelineDir, err)
	}
	if err := os.Mkdir(artifactsDir, 0777); err != nil {
		return fmt.Errorf("start pipeline build | create %s dir failed | err: %s", pipelineDir, err)
	}

	b.ArtifactsDir = artifactsDir

	return nil
}

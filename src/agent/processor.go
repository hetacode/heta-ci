package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/agent/structs"
)

type PipelineProcessor struct {
	pipelineTriggers *PipelineTriggers
	pipeline         *structs.Pipeline
	logChannel       chan string
	errorChannel     chan string
}

func NewPipelineProcessor(pipeline *structs.Pipeline, pt *PipelineTriggers) *PipelineProcessor {
	p := &PipelineProcessor{
		pipelineTriggers: pt,
		pipeline:         pipeline,
		logChannel:       make(chan string),
		errorChannel:     make(chan string),
	}
	p.parsePipelineForTriggersRegistration()

	return p
}

func (p *PipelineProcessor) Run() {
	p.logChannel <- fmt.Sprintf("run %s pipeline", p.pipeline.Name)

	for _, j := range p.pipeline.Jobs {
		p.executeJob(j)
	}
}

func (p *PipelineProcessor) Dispose() {
	close(p.errorChannel)
	close(p.logChannel)
}

func (p *PipelineProcessor) parsePipelineForTriggersRegistration() {
	for _, j := range p.pipeline.Jobs {
		p.pipelineTriggers.RegisterJob(j)
		for _, t := range j.Tasks {
			p.pipelineTriggers.RegisterTask(j.ID, t)
		}
	}
}

func (p *PipelineProcessor) executeJob(j structs.Job) {
	p.logChannel <- fmt.Sprintf("run '%s' job", j.DisplayName)

	pwd, _ := os.Getwd()
	pipelineTempDir := pwd + "/pipeline" // TODO: should be set up via cli parameter

	os.RemoveAll(pipelineTempDir)
	if err := os.Mkdir(pipelineTempDir, os.ModePerm); err != nil {
		p.errorChannel <- fmt.Sprintf("create pipeline temp directory err: %s", err)
		return
	}
	scriptsDir := pipelineTempDir + "/scripts"
	if err := os.Mkdir(scriptsDir, os.ModePerm); err != nil {
		p.errorChannel <- fmt.Sprintf("create scripts directory err: %s", err)
		return
	}
	defer os.RemoveAll(pipelineTempDir)

	c := NewContainer(j.Runner, pipelineTempDir)
	defer c.Dispose()

	var lastFailedTask *structs.Task
	var lastSuccessTask *structs.Task
	for _, t := range j.Tasks {
		// Task with conditions shouldn't be run in normal flow
		if len(t.Conditons) != 0 {
			continue
		}

		if err := p.executeTask(t, c, scriptsDir); err != nil {
			lastFailedTask = &t
			p.errorChannel <- err.Error()
			break
		} else {
			lastSuccessTask = &t
			break
		}
	}

	if lastFailedTask != nil {
		p.executeConditinonalTask(lastFailedTask, j.ID, c, scriptsDir, false)
	} else if lastSuccessTask != nil {
		p.executeConditinonalTask(lastSuccessTask, j.ID, c, scriptsDir, true)
	}

	p.logChannel <- fmt.Sprintf("job '%s' finished", j.DisplayName)
}

func (p *PipelineProcessor) executeConditinonalTask(t *structs.Task, jobID string, c *Container, scriptsDir string, onSuccess bool) bool {
	p.logChannel <- fmt.Sprint("conditional task")
	if t == nil {
		return false
	}

	conditionalTask := p.pipelineTriggers.GetTaskFor(*t, jobID, onSuccess)
	if conditionalTask == nil {
		return false
	}

	if err := p.executeTask(*conditionalTask, c, scriptsDir); err != nil {
		p.errorChannel <- err.Error()

		return p.executeConditinonalTask(conditionalTask, jobID, c, scriptsDir, false)
	} else {
		return p.executeConditinonalTask(conditionalTask, jobID, c, scriptsDir, true)
	}
}

func (p *PipelineProcessor) executeTask(t structs.Task, c *Container, scriptsDir string) error {
	p.logChannel <- fmt.Sprintf("run '%s' task", t.DisplayName)

	// Prepare script file
	uid, _ := uuid.GenerateUUID()
	filename := uid + ".sh"
	script := createScript(t.Command)
	f, err := os.Create(path.Join(scriptsDir, filename))
	if err != nil {
		return fmt.Errorf("execute task '%s' - create script err: %s", t.DisplayName, err)
	}
	_, err = f.Write(script)
	if err != nil {
		return fmt.Errorf("execute task '%s' - save script err: %s", t.DisplayName, err)

	}
	f.Chmod(775) // execute
	f.Close()

	// Execute script inside container
	if err := c.ExecuteScript(filename, p.logChannel); err != nil {
		return err
	}

	p.logChannel <- fmt.Sprintf("task '%s' done", t.DisplayName)
	return nil
}

func createScript(cmd []string) []byte {
	oneCmd := strings.Join(cmd, "\n")
	return []byte(oneCmd)
}

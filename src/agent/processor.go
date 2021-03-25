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
	haltChannel      chan struct{}

	pipelineHostDir  string
	jobScriptHostDir string
}

func NewPipelineProcessor(pipeline *structs.Pipeline, pt *PipelineTriggers, pipelineHostDir, scriptsHostDir string) *PipelineProcessor {
	p := &PipelineProcessor{
		pipelineTriggers: pt,
		pipeline:         pipeline,
		logChannel:       make(chan string),
		errorChannel:     make(chan string),
		haltChannel:      make(chan struct{}),
		pipelineHostDir:  pipelineHostDir,
		jobScriptHostDir: scriptsHostDir,
	}
	p.parsePipelineForTriggersRegistration()

	return p
}

func (p *PipelineProcessor) Run() {
	p.logChannel <- fmt.Sprintf("run %s pipeline", p.pipeline.Name)

	if err := os.Mkdir(p.jobScriptHostDir, os.ModePerm); err != nil {
		p.errorChannel <- fmt.Sprintf("create host scripts temp directory err: %s", err)
		return
	}
	if err := os.Mkdir(p.pipelineHostDir, os.ModePerm); err != nil {
		p.errorChannel <- fmt.Sprintf("create host pipeline temp directory err: %s", err)
		return
	}

	var lastFailedJob *structs.Job
	for _, j := range p.pipeline.Jobs {
		if len(j.Conditons) != 0 {
			continue
		}

		if err := p.executeJob(j); err != nil {
			lastFailedJob = &j
			break
		} else {
			if err := p.executeConditionalJob(&j, true); err != nil {
				p.haltChannel <- struct{}{}
				break
			}
		}
	}

	if lastFailedJob != nil {
		p.executeConditionalJob(lastFailedJob, false)
	}
	p.haltChannel <- struct{}{}
}

func (p *PipelineProcessor) Dispose() {
	close(p.errorChannel)
	close(p.logChannel)
	close(p.haltChannel)

	os.RemoveAll(p.pipelineHostDir)
	os.RemoveAll(p.jobScriptHostDir)
}

func (p *PipelineProcessor) executeJob(j structs.Job) error {
	p.logChannel <- fmt.Sprintf("run '%s' job", j.DisplayName)

	c := NewContainer(j.Runner, p.jobScriptHostDir, p.pipelineHostDir)
	defer c.Dispose()

	var lastFailedTask *structs.Task
	var lastFailedTaskErr error
	for _, t := range j.Tasks {
		// Task with conditions shouldn't be run in normal flow
		if len(t.Conditons) != 0 {
			continue
		}

		if err := p.executeTask(t, c, p.jobScriptHostDir); err != nil {
			lastFailedTask = &t
			lastFailedTaskErr = err
			p.errorChannel <- err.Error()
			break
		} else {
			if err := p.executeConditionalTask(&t, j.ID, c, p.jobScriptHostDir, true); err != nil {
				break
			}
		}
	}

	if lastFailedTask != nil {
		p.executeConditionalTask(lastFailedTask, j.ID, c, p.jobScriptHostDir, false)
		return lastFailedTaskErr
	}

	p.logChannel <- fmt.Sprintf("job '%s' finished", j.DisplayName)
	return nil
}

func (p *PipelineProcessor) executeConditionalJob(j *structs.Job, onSuccess bool) error {
	if j == nil {
		return nil
	}
	conditionalJob := p.pipelineTriggers.GetJobFor(*j, onSuccess)
	if conditionalJob == nil {
		return nil
	}

	if err := p.executeJob(*conditionalJob); err != nil {
		p.errorChannel <- fmt.Sprintf("job '%s' failed", conditionalJob.ID)
		p.executeConditionalJob(conditionalJob, false)
		return err
	} else {
		return p.executeConditionalJob(conditionalJob, true)
	}
}

func (p *PipelineProcessor) executeConditionalTask(t *structs.Task, jobID string, c *Container, scriptsDir string, onSuccess bool) error {
	if t == nil {
		return nil
	}

	conditionalTask := p.pipelineTriggers.GetTaskFor(*t, jobID, onSuccess)
	if conditionalTask == nil {
		return nil
	}

	if err := p.executeTask(*conditionalTask, c, scriptsDir); err != nil {
		p.errorChannel <- err.Error()

		p.executeConditionalTask(conditionalTask, jobID, c, scriptsDir, false)
		return err

	} else {
		return p.executeConditionalTask(conditionalTask, jobID, c, scriptsDir, true)
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

func (p *PipelineProcessor) parsePipelineForTriggersRegistration() {
	for _, j := range p.pipeline.Jobs {
		p.pipelineTriggers.RegisterJob(j)
		for _, t := range j.Tasks {
			p.pipelineTriggers.RegisterTask(j.ID, t)
		}
	}
}

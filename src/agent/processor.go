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
	pipeline     *structs.Pipeline
	logChannel   chan string
	errorChannel chan string
}

func NewPipelineProcessor(pipeline *structs.Pipeline) *PipelineProcessor {
	p := &PipelineProcessor{
		pipeline:     pipeline,
		logChannel:   make(chan string),
		errorChannel: make(chan string),
	}

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

func (p *PipelineProcessor) executeJob(j structs.Job) {
	p.logChannel <- fmt.Sprintf("run '%s' job", j.DisplayName)

	pwd, _ := os.Getwd()
	pipelineTempDir := pwd + "/pipeline" // TODO: should be set up via cli parameter

	os.RemoveAll(pipelineTempDir)
	if err := os.Mkdir(pipelineTempDir, os.ModePerm); err != nil {
		p.errorChannel <- fmt.Sprintf("create pipeline temp directory err: %s", err)
	}
	scriptsDir := pipelineTempDir + "/scripts"
	if err := os.Mkdir(scriptsDir, os.ModePerm); err != nil {
		p.errorChannel <- fmt.Sprintf("create scripts directory err: %s", err)
	}
	defer os.RemoveAll(pipelineTempDir)

	c := NewContainer(j.Runner, pipelineTempDir)
	defer c.Dispose()

	for _, t := range j.Tasks {
		if err := p.executeTask(t, c, scriptsDir); err != nil {
			c.Dispose()
			os.RemoveAll(pipelineTempDir)
			p.errorChannel <- err.Error()
			// TODO: in future check if any other task should be run on fail this one
		}
	}
	p.logChannel <- fmt.Sprintf("job '%s' finished", j.DisplayName)
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

package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/agent/structs"
)

type PipelineProcessor struct {
	pipeline   *structs.Pipeline
	logChannel chan string
}

func NewPipelineProcessor(pipeline *structs.Pipeline) *PipelineProcessor {
	p := &PipelineProcessor{
		pipeline:   pipeline,
		logChannel: make(chan string),
	}

	return p
}

func (p *PipelineProcessor) Run() {
	p.logChannel <- fmt.Sprintf("run %s pipeline", p.pipeline.Name)
	defer close(p.logChannel)

	for _, j := range p.pipeline.Jobs {
		p.executeJob(j)
	}
}

func (p *PipelineProcessor) executeJob(j structs.Job) {
	p.logChannel <- fmt.Sprintf("run %s job", j.Name)

	pwd, _ := os.Getwd()
	pipelineTempDir := pwd + "/pipeline" // TODO: should be set up via cli parameter

	os.RemoveAll(pipelineTempDir)
	if err := os.Mkdir(pipelineTempDir, os.ModePerm); err != nil {
		log.Fatalf("create pipeline temp directory err: %s", err)
	}
	scriptsDir := pipelineTempDir + "/scripts"
	if err := os.Mkdir(scriptsDir, os.ModePerm); err != nil {
		log.Fatalf("create scripts directory err: %s", err)
	}
	defer os.RemoveAll(pipelineTempDir)

	c := NewContainer(j.Runner, pipelineTempDir)
	defer c.Dispose()

	for _, t := range j.Tasks {
		p.executeTask(t, c, scriptsDir)
	}
}

func (p *PipelineProcessor) executeTask(t structs.Task, c *Container, scriptsDir string) {
	p.logChannel <- fmt.Sprintf("run %s task", t.Name)

	// Prepare script file
	uid, _ := uuid.GenerateUUID()
	filename := uid + ".sh"
	script := createScript(t.Command)
	f, err := os.Create(path.Join(scriptsDir, filename))
	if err != nil {
		log.Fatalf("execute task '%s' - create script err: %s", t.Name, err)
	}
	_, err = f.Write(script)
	if err != nil {
		log.Fatalf("execute task '%s' - save script err: %s", t.Name, err)
	}
	f.Chmod(775)
	f.Close()

	// Execute script inside container
	if err := c.ExecuteScript(filename, p.logChannel); err != nil {
		log.Fatal(err)
	}

	p.logChannel <- fmt.Sprintf("task %s done", t.Name)

}

func createScript(cmd []string) []byte {
	oneCmd := strings.Join(cmd, "\n")
	return []byte(oneCmd)
}

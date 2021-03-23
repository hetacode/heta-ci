package main

import (
	"fmt"

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

	for _, j := range p.pipeline.Jobs {
		p.executeJob(j)
	}
}

func (p *PipelineProcessor) executeJob(j structs.Job) {
	p.logChannel <- fmt.Sprintf("run %s job", j.Name)

	c := NewContainer(j.Runner)
	defer c.Dispose()

	for _, t := range j.Tasks {
		p.executeTask(t, c)
	}
}

func (p *PipelineProcessor) executeTask(t structs.Task, c *Container) {
	p.logChannel <- fmt.Sprintf("run %s task", t.Name)

}

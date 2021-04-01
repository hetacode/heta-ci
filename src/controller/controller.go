package main

import "github.com/hetacode/heta-ci/structs"

type Controller struct {
	pipelines []*structs.Pipeline
	workers   []*PipelineWorker
	agents    []*Agent // list of free agents
}

func NewController() *Controller {
	c := &Controller{}
	return c
}

func (c *Controller) AddPipeline(p *structs.Pipeline) {
	c.pipelines = append(c.pipelines, p)
}

func (c *Controller) Execute() {
	// TODO: a correct way - it should iterate through git repositories
	for _, p := range c.pipelines {
		w := NewPipelineWorker(p)
		go w.Run()
	}
}

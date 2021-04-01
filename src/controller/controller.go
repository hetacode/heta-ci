package main

import "github.com/hetacode/heta-ci/structs"

type PipelineStatus string

type Controller struct {
	pipelines []*structs.Pipeline
}

func NewController() *Controller {
	c := &Controller{}
	return c
}

func (c *Controller) AddPipeline(p *structs.Pipeline) {
	c.pipelines = append(c.pipelines, p)
}

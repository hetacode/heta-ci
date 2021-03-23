package main

import (
	"log"
	"time"

	"github.com/hetacode/heta-ci/agent/structs"
)

func main() {
	p := NewPipelineProcessor(preparePipeline())
	go p.Run()

	t := time.NewTimer(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case logStr, more := <-p.logChannel:
			if !more {
				return
			}
			log.Print(logStr)
		case <-t.C:
			return
		}
	}
}

func preparePipeline() *structs.Pipeline {
	pipeline := &structs.Pipeline{
		Name: "Test shell scripts in one container",
		Jobs: []structs.Job{
			{
				Name:   "Container job",
				Runner: "ubuntu:20.10",
				Tasks: []structs.Task{
					{
						Name: "Correct script",
						Command: []string{
							"echo Start",
							"cd /etc && ls -al",
							"echo End",
						},
					},
					{
						Name: "Failed script",
						Command: []string{
							"echo Start",
							"cd /etc && lt -al",
							"echo End",
						},
					},
				},
			},
		},
	}

	return pipeline
}

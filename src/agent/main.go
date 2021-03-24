package main

import (
	"log"
	"time"

	"github.com/hetacode/heta-ci/agent/structs"
)

func main() {
	timeoutCh := make(chan struct{})
	defer close(timeoutCh)
	p := NewPipelineProcessor(preparePipeline())
	defer p.Dispose()

	go p.Run()

	go func() {
		t := time.NewTimer(time.Minute * 3)
		<-t.C
		timeoutCh <- struct{}{}
	}()

	isRunning := true
	for {
		if !isRunning {
			break
		}
		select {
		case logStr, more := <-p.logChannel:
			if !more {
				isRunning = false
			}
			log.Print(logStr)
		case errorStr := <-p.errorChannel:
			log.Printf("\033[32mError: %s\033[0m", errorStr)
			isRunning = false
		case <-timeoutCh:
			isRunning = false
		}
	}

	log.Println("pipeline finished")
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

package main

import (
	"log"
	"time"

	"github.com/hetacode/heta-ci/agent/structs"
)

func main() {
	timeoutCh := make(chan struct{})
	defer close(timeoutCh)
	pt := NewPipelineTriggers()
	p := NewPipelineProcessor(preparePipeline(), pt)
	defer p.Dispose()

	go p.Run()

	go func() {
		t := time.NewTimer(time.Second * 30)
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
			log.Print(logStr)
			if !more {
				isRunning = false
			}
		case errorStr, more := <-p.errorChannel:
			log.Printf("\033[31mError: %s\033[0m", errorStr)
			if !more {
				isRunning = false
			}
		case <-p.haltChannel:
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
				ID:          "test_alpine",
				DisplayName: "Alpine runner",
				Runner:      "alpine",
				Tasks: []structs.Task{
					{
						ID:          "correct",
						DisplayName: "Correct script",
						Command: []string{
							"echo Start",
							"cd /etc && ls -al",
							"echo End",
						},
					},
				},
			},
			// {
			// 	ID:          "test_busybox",
			// 	DisplayName: "Busybox runner",
			// 	Runner:      "busybox",
			// 	Tasks: []structs.Task{
			// 		{
			// 			ID:          "correct",
			// 			DisplayName: "Correct script",
			// 			Command: []string{
			// 				"echo Start",
			// 				"cd /etc && ls -al",
			// 				"echo End",
			// 			},
			// 		},
			// 	},
			// },
			// {
			// 	ID:          "test_ubuntu",
			// 	DisplayName: "Ubuntu runner ",
			// 	Runner:      "ubuntu:20.04",
			// 	Tasks: []structs.Task{
			// 		{
			// 			ID:          "correct",
			// 			DisplayName: "Correct script",
			// 			Command: []string{
			// 				"echo Start",
			// 				"cd /etc && ls -al",
			// 				"echo End",
			// 			},
			// 		},
			// 	},
			// },
			// {
			// 	ID:          "test_arch",
			// 	DisplayName: "Arch runner ",
			// 	Runner:      "archlinux",
			// 	Tasks: []structs.Task{
			// 		{
			// 			ID:          "correct",
			// 			DisplayName: "Correct script",
			// 			Command: []string{
			// 				"echo Start",
			// 				"cd /etc && ls -al",
			// 				"echo End",
			// 			},
			// 		},
			// 	},

			// {
			// 	ID:          "fail",
			// 	DisplayName: "Failed script",
			// 	Command: []string{
			// 		"echo Start",
			// 		"cd /etc && lt -al",
			// 		"echo End",
			// 	},
			// },
			// {
			// 	ID:          "on_success_correct_task",
			// 	DisplayName: "Launch when 'test' task finish successfuly",
			// 	Conditons: []structs.Conditon{
			// 		{
			// 			Type: structs.OnSuccess,
			// 			On:   "correct",
			// 		},
			// 	},
			// 	Command: []string{
			// 		"apt update && apt install -y figlet",
			// 		"figlet Success",
			// 	},
			// },
			// },
			// },
			// {
			// 	ID:          "when_test_failed",
			// 	DisplayName: "Run conditionaly after test failed",
			// 	Runner:      "ubuntu:20.10",
			// 	Conditons: []structs.Conditon{
			// 		{
			// 			Type: structs.OnFailure,
			// 			On:   "test",
			// 		},
			// 	},
			// 	Tasks: []structs.Task{
			// 		{
			// 			ID:          "message",
			// 			DisplayName: "Message task for job 2",
			// 			Command: []string{
			// 				"apt update && apt install -y figlet",
			// 				"figlet \"Don't worry!\"",
			// 			},
			// 		},
			// 	},
			// },
		},
	}

	return pipeline
}

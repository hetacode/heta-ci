package main

import (
	"log"
	"net"

	proto "github.com/hetacode/heta-ci/proto"
	"github.com/hetacode/heta-ci/structs"
	"google.golang.org/grpc"
)

func main() {
	logCh := make(chan string)
	errLogCh := make(chan string)
	c := NewController()
	c.AddPipeline(preparePipeline())

	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Panic(err)
	}
	srv := grpc.NewServer()
	cs := NewCommunicationServer(logCh, errLogCh)
	proto.RegisterCommunicationServer(srv, cs)

	go func() {
		for {
			select {
			case logStr := <-logCh:
				log.Printf("\033[97m%s\033[0m", logStr)
			case errLogStr := <-errLogCh:
				log.Printf("\033[31m%s\033[0m", errLogStr)
			}

		}
	}()

	err = srv.Serve(lis)
	if err != nil {
		log.Panic(err)
	}
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
					{
						ID:          "correct",
						DisplayName: "Correct script",
						Command: []string{
							"echo Start",
							"echo job artifacts dir: $AGENT_JOB_ARTIFACTS_DIR",
							"echo scripts dir: $AGENT_SCRIPTS_DIR",
							"echo End",
						},
					},
				},
			},
		},
	}

	return pipeline
}

package main

import (
	"log"
	"net"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/controller/eventhandlers"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/hetacode/heta-ci/events/agent"
	proto "github.com/hetacode/heta-ci/proto"
	"github.com/hetacode/heta-ci/structs"
	"google.golang.org/grpc"
)

// TODO:
// + 1. prepare events for both side - like Controller -> Agent: AssignedAgentID (just after connect to controller)
// + 2. prepare whole stuff for event handling - eventsmapper, eventshandlers, connect events with eventshandlers
// - 3. added grpc client on the agent side

func main() {

	c := utils.NewController()
	ehm := registerEventHandlers(c)

	c.AddPipeline(preparePipeline())

	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Panic(err)
	}
	srv := grpc.NewServer()
	cs := utils.NewCommunicationServer(ehm)
	proto.RegisterCommunicationServer(srv, cs)

	err = srv.Serve(lis)
	if err != nil {
		log.Panic(err)
	}
}

func registerEventHandlers(c *utils.Controller) *goeh.EventsHandlerManager {
	ehm := goeh.NewEventsHandlerManager()
	ehm.Register(new(agent.LogMessageEvent), &eventhandlers.LogMessageEventHandler{Controller: c})
	ehm.Register(new(agent.JobFinishedEvent), &eventhandlers.JobFinishedEventHandler{Controller: c})

	return ehm
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

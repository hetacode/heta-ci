package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/controller/eventhandlers"
	"github.com/hetacode/heta-ci/controller/handlers"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/hetacode/heta-ci/events/agent"
	proto "github.com/hetacode/heta-ci/proto"
	"github.com/hetacode/heta-ci/structs"
	"google.golang.org/grpc"
)

func main() {
	addAgentCh := make(chan *utils.Agent)
	removeAgentCh := make(chan *utils.Agent)

	c := utils.NewController(addAgentCh, removeAgentCh)
	ehm := registerEventHandlers(c)

	c.AddPipeline(preparePipeline())

	go initRestApi()
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Panic(err)
	}
	srv := grpc.NewServer()
	cs := utils.NewCommunicationServer(ehm, addAgentCh, removeAgentCh)
	proto.RegisterCommunicationServer(srv, cs)

	// TEST PIPELINE EXECUTIONS
	go func() {
		time.Sleep(10 * time.Second)
		c.Execute()
	}()

	err = srv.Serve(lis)
	if err != nil {
		log.Panic(err)
	}
}

func initRestApi() {
	h := &handlers.Handlers{}
	r := mux.NewRouter()
	r.HandleFunc("/download/{category}/{buildId}", h.DownloadFileHandler)
	r.HandleFunc("/upload/{buildId}/{jobId}", h.UploadArtifactsHandler)
	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:5080",
	}
	srv.ListenAndServe()
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
						ID:          "correct_ls",
						DisplayName: "Correct script - ls dir",
						Command: []string{
							"echo Start",
							"cd /etc && ls -al",
							"echo End",
						},
					},
					{
						ID:          "correct_env",
						DisplayName: "Correct script - env",
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

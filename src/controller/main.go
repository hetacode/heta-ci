package main

import (
	"log"
	"net"
	"net/http"
	"os"
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
	"gopkg.in/yaml.v2"
)

func main() {
	addAgentCh := make(chan *utils.Agent)
	removeAgentCh := make(chan *utils.Agent)

	c := utils.NewController(addAgentCh, removeAgentCh)
	ehm := registerEventHandlers(c)

	c.AddPipeline(preparePipeline())

	go initRestApi(c)
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

func initRestApi(c *utils.Controller) {
	h := &handlers.Handlers{Controller: c}
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
	b, err := os.ReadFile(".heta-ci/pipeline.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var pipeline *structs.Pipeline
	if err := yaml.Unmarshal(b, &pipeline); err != nil {
		log.Fatal(err)
	}

	return pipeline
	// // pipeline := &structs.Pipeline{
	// // 	Name: "Test shell scripts in one container",
	// // 	Jobs: []structs.Job{
	// // 		{
	// // 			ID:          "test_alpine",
	// // 			DisplayName: "Alpine runner",
	// // 			Runner:      "alpine",
	// // 			Tasks: []structs.Task{
	// // 				{
	// // 					ID:          "correct_ls",
	// // 					DisplayName: "Correct script - ls dir",
	// // 					Command: []string{
	// // 						"echo Start",
	// // 						"uname -a >> $AGENT_TASKS_DIR/uname.txt",
	// // 						"echo End",
	// // 					},
	// // 				},
	// // 				{
	// // 					ID:          "correct_env",
	// // 					DisplayName: "Correct script - env",
	// // 					Command: []string{
	// // 						"echo Start",
	// // 						"echo job artifacts IN dir: $AGENT_JOB_ARTIFACTS_IN_DIR",
	// // 						"echo job artifacts OUT dir: $AGENT_JOB_ARTIFACTS_OUT_DIR",
	// // 						"echo job tasks dir: $AGENT_TASKS_DIR",
	// // 						"echo scripts dir: $AGENT_SCRIPTS_DIR",
	// // 						"echo End",
	// // 					},
	// // 				},
	// // 				{
	// // 					ID:          "read_uname_file",
	// // 					DisplayName: "Read the file with uname saved value",
	// // 					Command: []string{
	// // 						"echo Start",
	// // 						"cat $AGENT_TASKS_DIR/uname.txt",
	// // 						"mkdir $AGENT_JOB_ARTIFACTS_OUT_DIR/test",
	// // 						"echo 'lorem ipsum' >>  $AGENT_JOB_ARTIFACTS_OUT_DIR/test/lorem.txt",
	// // 						"cp $AGENT_TASKS_DIR/uname.txt $AGENT_JOB_ARTIFACTS_OUT_DIR/",
	// // 						"echo end",
	// // 					},
	// // 				},
	// // 			},
	// // 		},
	// // 		{
	// // 			ID:          "test_busybox",
	// // 			DisplayName: "Busybox runner",
	// // 			Runner:      "busybox",
	// // 			Tasks: []structs.Task{
	// // 				{
	// // 					ID:          "correct",
	// // 					DisplayName: "Correct script",
	// // 					Command: []string{
	// // 						"echo Start",
	// // 						"ls -la $AGENT_JOB_ARTIFACTS_IN_DIR/",
	// // 						"cp -r $AGENT_JOB_ARTIFACTS_IN_DIR/* $AGENT_TASKS_DIR/",
	// // 						"echo End",
	// // 					},
	// // 				},
	// // 				{
	// // 					ID:          "correct",
	// // 					DisplayName: "Read uname file",
	// // 					Command: []string{
	// // 						"echo Start",
	// // 						"cat $AGENT_TASKS_DIR/uname.txt",
	// // 						"echo End",
	// // 					},
	// // 				},
	// // 			},
	// // 		},
	// // 		// {
	// // 		// 	ID:          "when_test_busybox_failed",
	// // 		// 	DisplayName: "Run conditionaly after test busybox failed",
	// // 		// 	Runner:      "ubuntu:20.10",
	// // 		// 	Conditons: []structs.Conditon{
	// // 		// 		{
	// // 		// 			Type: structs.OnFailure,
	// // 		// 			On:   "test_busybox",
	// // 		// 		},
	// // 		// 	},
	// // 		// 	Tasks: []structs.Task{
	// // 		// 		{
	// // 		// 			ID:          "message",
	// // 		// 			DisplayName: "Message task for job 2",
	// // 		// 			Command: []string{
	// // 		// 				"apt update && apt install -y figlet",
	// // 		// 				"figlet \"Don't worry!\"",
	// // 		// 			},
	// // 		// 		},
	// // 		// 	},
	// // 		// },
	// // 	},
	// }
}

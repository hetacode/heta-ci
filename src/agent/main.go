package main

import (
	"context"
	"log"
	"time"

	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/agent/eventhandlers"
	"github.com/hetacode/heta-ci/agent/handlers"
	"github.com/hetacode/heta-ci/events/controller"
	"github.com/hetacode/heta-ci/proto"
	"google.golang.org/grpc"
)

func main() {

	a := app.NewApp()

	timeoutCh := make(chan struct{})
	defer close(timeoutCh)
	// pe := NewPipelineEnvironments(ScriptsDir, JobDir, PipelineDir)
	// pt := NewPipelineTriggers()
	// p := NewPipelineProcessor(preparePipeline(), pt, pe, pipelineHostDir, scriptsHostDir)
	// defer p.Dispose()
	//
	// go p.Run()

	go func() {
		t := time.NewTimer(time.Second * 30)
		<-t.C
		timeoutCh <- struct{}{}
	}()

	con, err := grpc.Dial(":5000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("agent | cannot connect to the controller | err: %s", err)
	}
	defer con.Close()

	client := proto.NewCommunicationClient(con)
	stream, err := client.MessagingService(context.Background())
	if err != nil {
		log.Fatalf("agent | failed to call messaging service | err: %+v", err)
	}

	ms := handlers.NewMessagingServiceHandler(a.Config, stream)
	a.MessagingService = ms
	go ms.ReceivingMessages()

	log.Println("pipeline finished")
}

func registerEventHandlers(a *app.App) {
	a.EventsHandlerManager.Register(new(controller.AgentConfirmedEvent), &eventhandlers.AgentConfirmedEventHandler{App: a})
	a.EventsHandlerManager.Register(new(controller.StartJobCommand), &eventhandlers.StartJobCommandHandler{App: a})
}

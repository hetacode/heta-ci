package main

import (
	"context"
	"log"
	"os"

	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/agent/eventhandlers"
	"github.com/hetacode/heta-ci/agent/handlers"
	"github.com/hetacode/heta-ci/events/controller"
	"github.com/hetacode/heta-ci/proto"
	"google.golang.org/grpc"
)

func main() {
	wait := make(chan bool)
	a := app.NewApp()
	registerEventHandlers(a)

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

	ms := handlers.NewMessagingServiceHandler(a.Config, a.EventsHandlerManager, stream)
	a.MessagingService = ms
	go ms.ReceivingMessages()

	h, _ := os.Hostname()
	log.Printf("agent %s is running", h)

	<-wait
}

func registerEventHandlers(a *app.App) {
	a.EventsHandlerManager.Register(new(controller.AgentConfirmedEvent), &eventhandlers.AgentConfirmedEventHandler{App: a})
	a.EventsHandlerManager.Register(new(controller.StartJobCommand), &eventhandlers.StartJobCommandHandler{App: a})
}

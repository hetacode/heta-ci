package main

import (
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/proto"
)

type MessagingServiceHandler struct {
	app                  *app.App
	stream               proto.Communication_MessagingServiceClient
	eventsHandlerManager *goeh.EventsHandlerManager
}

func NewMessagingServiceHandler(app *app.App, stream proto.Communication_MessagingServiceClient) *MessagingServiceHandler {
	ms := &MessagingServiceHandler{
		app:    app,
		stream: stream,
	}

	return ms
}

func (s *MessagingServiceHandler) ReceivingMessages() {
	for {
		msg, err := s.stream.Recv()
		if err != nil {
			log.Fatalf("agent | reveive message | err: %s", err)
		}

		// TODO:
		// call events handler manager

		ev, err := s.app.Config.EventsMapper.Resolve(msg.Payload)
		if err != nil {
			log.Printf("agent | failed resolve message type: %s | err: %s", msg.Payload, err)
			return
		}
		s.eventsHandlerManager.Execute(ev)
	}
}

func (s *MessagingServiceHandler) SendMessage(e goeh.Event) {
	mes := &proto.MessageFromAgent{
		Id:       s.app.Config.AgentID,
		Hostname: s.app.Config.Hostname,
		Type:     e.GetType(),
		Payload:  e.GetPayload(),
	}
	if err := s.stream.Send(mes); err != nil {
		log.Fatalf("agent | send message | err: %s", err)
	}
}

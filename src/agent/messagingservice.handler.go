package main

import (
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/proto"
)

type MessagingServiceHandler struct {
	config *Config
	stream proto.Communication_MessagingServiceClient
}

func NewMessagingServiceHandler(config *Config, stream proto.Communication_MessagingServiceClient) *MessagingServiceHandler {
	ms := &MessagingServiceHandler{
		config: config,
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
		log.Fatalf("Unimplemnted ReceivingMessages: %+v", msg)
	}
}

func (s *MessagingServiceHandler) SendMessage(e goeh.Event) {
	mes := &proto.MessageFromAgent{
		Id:       s.config.AgentID,
		Hostname: s.config.Hostname,
		Type:     e.GetType(),
		Payload:  e.GetPayload(),
	}
	if err := s.stream.Send(mes); err != nil {
		log.Fatalf("agent | send message | err: %s", err)
	}
}

package handlers

import (
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/utils"
	"github.com/hetacode/heta-ci/proto"
)

type MessagingServiceHandler struct {
	config               *utils.Config
	stream               proto.Communication_MessagingServiceClient
	eventsHandlerManager *goeh.EventsHandlerManager
}

func NewMessagingServiceHandler(config *utils.Config, eventsHandlerManager *goeh.EventsHandlerManager, stream proto.Communication_MessagingServiceClient) *MessagingServiceHandler {
	ms := &MessagingServiceHandler{
		config:               config,
		stream:               stream,
		eventsHandlerManager: eventsHandlerManager,
	}

	return ms
}

func (s *MessagingServiceHandler) ReceivingMessages() {
	for {
		msg, err := s.stream.Recv()
		if err != nil {
			log.Fatalf("agent | receive message | err: %s", err)
		}

		ev, err := s.config.EventsMapper.Resolve(msg.Payload)
		if err != nil {
			log.Printf("agent | failed resolve message type: %+v | err: %s", msg, err)
			return
		}
		s.eventsHandlerManager.Execute(ev)
	}
}

func (s *MessagingServiceHandler) SendMessage(e goeh.Event) {
	e.SavePayload(e)
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

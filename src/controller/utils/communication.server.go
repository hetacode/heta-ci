package utils

import (
	"log"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
	evcontroller "github.com/hetacode/heta-ci/events/controller"
	proto "github.com/hetacode/heta-ci/proto"
)

type CommunicationServer struct {
	proto.UnimplementedCommunicationServer
	AgentErrorChan chan AgentError
	Agents         []*Agent
}

func NewCommunicationServer() *CommunicationServer {
	errCh := make(chan AgentError)
	c := &CommunicationServer{
		AgentErrorChan: errCh,
	}

	return c
}

func (s *CommunicationServer) MessagingService(client proto.Communication_MessagingServiceServer) error {

	a := NewAgent(client, s.AgentErrorChan)
	go a.ReceivingMessages()

	s.Agents = append(s.Agents, a)

	euid, _ := uuid.GenerateUUID()
	ev := &evcontroller.AgentConfirmedEvent{
		EventData: &goeh.EventData{ID: euid},
		AgentID:   a.ID,
	}
	a.SendMessage(ev)

	return nil
}

func (s *CommunicationServer) AgentErrorsReceiver() {
	for err := range s.AgentErrorChan {
		for i, a := range s.Agents {
			if a.ID == err.ID {
				s.Agents = append(s.Agents[:i], s.Agents[i+1:]...)
				break
			}
		}

		log.Printf("\033[31m%s\033[0m", err.Error())
	}
}

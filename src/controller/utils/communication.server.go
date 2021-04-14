package utils

import (
	"log"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/events"
	evcontroller "github.com/hetacode/heta-ci/events/controller"
	proto "github.com/hetacode/heta-ci/proto"
)

type CommunicationServer struct {
	proto.UnimplementedCommunicationServer
	EventsHandlerManager *goeh.EventsHandlerManager
	AgentErrorChan       chan AgentError
	Agents               []*Agent
	eventsMapper         *goeh.EventsMapper
	addAgentChan         chan *Agent
	removeAgentChan      chan *Agent
}

func NewCommunicationServer(ehm *goeh.EventsHandlerManager, addAgentCh, removeAgentCh chan *Agent) *CommunicationServer {
	errCh := make(chan AgentError)
	c := &CommunicationServer{
		AgentErrorChan:       errCh,
		EventsHandlerManager: ehm,
		eventsMapper:         events.NewEventsMapper(),
		addAgentChan:         addAgentCh,
		removeAgentChan:      removeAgentCh,
	}

	return c
}

func (s *CommunicationServer) MessagingService(client proto.Communication_MessagingServiceServer) error {
	a := NewAgent(client, s.AgentErrorChan, s.EventsHandlerManager)
	go a.ReceivingMessages(s.eventsMapper)

	s.Agents = append(s.Agents, a)

	euid, _ := uuid.GenerateUUID()
	ev := &evcontroller.AgentConfirmedEvent{
		EventData: &goeh.EventData{ID: euid},
		AgentID:   a.ID,
	}
	go a.SendMessage(ev)
	log.Printf("agent %s connected", a.ID)
	s.addAgentChan <- a

	err := <-s.AgentErrorChan
	log.Printf("\033[31m%s\033[0m", err.Error())

	for i, a := range s.Agents {
		if a.ID == err.ID {
			s.Agents = append(s.Agents[:i], s.Agents[i+1:]...)
			break
		}
	}
	s.removeAgentChan <- a

	return nil
}

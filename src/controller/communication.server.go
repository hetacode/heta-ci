package main

import proto "github.com/hetacode/heta-ci/proto"

type CommunicationServer struct {
	proto.UnimplementedCommunicationServer
	AgentErrorChan chan AgentError
	Agents         []*Agent
	LogChan        chan string
	ErrLogChan     chan string
}

func NewCommunicationServer(logCh, errLogCh chan string) *CommunicationServer {
	errCh := make(chan AgentError)
	c := &CommunicationServer{
		AgentErrorChan: errCh,
		LogChan:        logCh,
		ErrLogChan:     errLogCh,
	}

	return c
}

func (s *CommunicationServer) MessagingService(client proto.Communication_MessagingServiceServer) error {

	agent := NewAgent(client, s.AgentErrorChan)
	go agent.ReceivingMessages()

	s.Agents = append(s.Agents, agent)

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

		s.ErrLogChan <- err.Error()
	}
}

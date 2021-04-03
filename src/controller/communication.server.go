package main

import proto "github.com/hetacode/heta-ci/proto"

type CommunicationServer struct {
	proto.UnimplementedCommunicationServer
	ErrorChan chan AgentError
}

func NewCommunicationServer() *CommunicationServer {
	errCh := make(chan AgentError)
	c := &CommunicationServer{
		ErrorChan: errCh,
	}

	// TODO: errors receiver

	return c
}

func (s *CommunicationServer) MessagingService(client proto.Communication_MessagingServiceServer) error {

	agent := NewAgent(client, s.ErrorChan)
	go agent.ReceivingMessages()

	return nil
}

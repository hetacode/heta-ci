package main

import proto "github.com/hetacode/heta-ci/proto"

type CommunicationServer struct {
	proto.UnimplementedCommunicationServer
}

func (s *CommunicationServer) MessagingService(client proto.Communication_MessagingServiceServer) error {

	panic("Not implemented")
}

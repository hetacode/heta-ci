package utils

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
	proto "github.com/hetacode/heta-ci/proto"
)

type AgentError struct {
	ID    string
	error string
}

func NewAgentError(id, message string) AgentError {
	return AgentError{
		ID:    id,
		error: message,
	}
}

func (a AgentError) Error() string {
	return a.error
}

type Agent struct {
	ID        string
	Client    proto.Communication_MessagingServiceServer
	ErrorChan chan AgentError
}

func NewAgent(client proto.Communication_MessagingServiceServer, errCh chan AgentError) *Agent {
	uid, _ := uuid.GenerateUUID()
	a := &Agent{
		ID:        uid,
		Client:    client,
		ErrorChan: errCh,
	}

	return a
}

func (a *Agent) ReceivingMessages() {
	for {
		msg, err := a.Client.Recv()
		if err != nil {
			a.ErrorChan <- NewAgentError(a.ID, fmt.Sprintf("agent: %s receive err: %s", a.ID, err))
			return
		}

		log.Printf("agent: %s msg: %s", a.ID, msg.String())
	}
}

func (a *Agent) SendMessage(event goeh.Event) {
	send := &proto.MessageFromController{
		Type:    event.GetType(),
		Payload: event.GetPayload(),
	}
	err := a.Client.Send(send)
	if err != nil {
		a.ErrorChan <- NewAgentError(a.ID, fmt.Sprintf("agent: %s send err: %s", a.ID, err))
		return
	}
}

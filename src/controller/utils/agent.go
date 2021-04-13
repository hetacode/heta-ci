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
	ID                   string
	client               proto.Communication_MessagingServiceServer
	errorChan            chan AgentError
	eventsHandlerManager *goeh.EventsHandlerManager
}

func NewAgent(client proto.Communication_MessagingServiceServer, errCh chan AgentError, ehm *goeh.EventsHandlerManager) *Agent {
	uid, _ := uuid.GenerateUUID()
	a := &Agent{
		ID:                   uid,
		client:               client,
		errorChan:            errCh,
		eventsHandlerManager: ehm,
	}

	return a
}

func (a *Agent) ReceivingMessages(em *goeh.EventsMapper) {
	for {
		msg, err := a.client.Recv()
		if err != nil {
			a.errorChan <- NewAgentError(a.ID, fmt.Sprintf("agent: %s receive err: %s", a.ID, err))
			return
		}

		log.Printf("agent: %s msg: %s", a.ID, msg.String())

		ev, err := em.Resolve(msg.Payload)
		if err != nil {
			a.errorChan <- NewAgentError(a.ID, fmt.Sprintf("agent: %s cannot regnize event err: %s", a.ID, err))
			return
		}
		a.eventsHandlerManager.Execute(ev)
	}
}

func (a *Agent) SendMessage(event goeh.Event) {
	event.SavePayload(event)
	send := &proto.MessageFromController{
		Type:    event.GetType(),
		Payload: event.GetPayload(),
	}

	err := a.client.Send(send)
	if err != nil {
		a.errorChan <- NewAgentError(a.ID, fmt.Sprintf("agent: %s send err: %s", a.ID, err))
		return
	}
}

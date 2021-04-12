package events

import (
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/events/agent"
	"github.com/hetacode/heta-ci/events/controller"
)

func NewEventsMapper() *goeh.EventsMapper {
	e := new(goeh.EventsMapper)
	e.Register(new(agent.JobFinishedEvent))
	e.Register(new(agent.LogMessageEvent))
	e.Register(new(controller.AgentConfirmedEvent))
	e.Register(new(controller.StartJobCommand))
	return e
}

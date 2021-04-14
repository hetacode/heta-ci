package eventhandlers

import (
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/app"
	events "github.com/hetacode/heta-ci/events/controller"
)

type AgentConfirmedEventHandler struct {
	App *app.App
}

func (h *AgentConfirmedEventHandler) Handle(event goeh.Event) {
	ev := event.(*events.AgentConfirmedEvent)

	log.Printf("agent | agent confirmed - id: %s", ev.AgentID)
	h.App.Config.AgentID = ev.AgentID
}

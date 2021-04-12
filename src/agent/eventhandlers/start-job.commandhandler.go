package eventhandlers

import (
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/events/controller"
)

type StartJobCommandHandler struct {
	App *app.App
}

func (h *StartJobCommandHandler) Handle(event goeh.Event) {
	ev := event.(*controller.StartJobCommand)
	log.Fatalf("unimplemented StartJobCommandHandler %+v", ev)
}

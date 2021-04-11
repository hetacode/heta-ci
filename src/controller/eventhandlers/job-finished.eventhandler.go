package eventhandlers

import (
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/hetacode/heta-ci/events/agent"
)

type JobFinishedEventHandler struct {
	Controller *utils.Controller
}

func (e *JobFinishedEventHandler) Handle(event goeh.Event) {
	ev := event.(*agent.JobFinishedEvent)
	b, ok := e.Controller.Builds[ev.BuildID]
	if !ok {
		log.Printf("JobFinishedEvent | cannot find build id: %s", ev.BuildID)
		return
	}

	// TODO:
	// to implement jobs flow - from agent
	log.Fatalf("Unimplemented JobFinishedEventHandler %+v", b)
}

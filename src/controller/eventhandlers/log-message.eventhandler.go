package eventhandlers

import (
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/controller/app"
	"github.com/hetacode/heta-ci/events/agent"
)

type LogMessageEventHandler struct {
	Controller *app.Controller
}

func (e *LogMessageEventHandler) Handle(event goeh.Event) {
	ev := event.(*agent.LogMessageEvent)
	b, ok := e.Controller.Builds[ev.BuildID]
	if !ok {
		log.Printf("LogMessageEvent | cannot find build id: %s", ev.BuildID)
		return
	}

	switch ev.LogType {
	case agent.InfoLogType:
		b.LogChan <- ev.Message
	case agent.ErrorLogType:
		b.ErrLogChan <- ev.Message
	}
}

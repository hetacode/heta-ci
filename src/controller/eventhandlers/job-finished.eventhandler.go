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
	e.Controller.ReturnAgentCh <- b.Agent

	switch ev.Reason {
	case agent.ErrorJobFinishReason:
		b.ErrLogChan <- ev.Message

		nextJob := b.Triggers.GetJobFor(ev.JobID, false)
		if nextJob == nil {
			return
		}
		b.StartJob(nextJob, true)
		return
	case agent.CompleteJobFinishReason:
		b.LogChan <- ev.Message

		nextJob := b.Triggers.GetJobFor(ev.JobID, true)
		if nextJob != nil {
			b.StartJob(nextJob, true)
			return
		}
		// Go to looking for another job
	}

	if ev.WasConditionalJob {
		// If finished job was executed in conditional flow
		// we shouldn't never go to normal iterating path
		// It exactly means end of pipeline!
		return
	}

	start := false
	for _, j := range b.Pipeline.Jobs {
		if ev.JobID == j.ID && !start {
			start = true
			continue
		}

		if !start {
			continue
		}

		if len(j.Conditons) != 0 {
			continue
		}

		b.StartJob(&j, false)
		return
	}
}

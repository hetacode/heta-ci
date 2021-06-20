package eventhandlers

import (
	"fmt"
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/controller/app"
	"github.com/hetacode/heta-ci/controller/enums"
	"github.com/hetacode/heta-ci/events/agent"
)

type JobFinishedEventHandler struct {
	Controller *app.Controller
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
			if err := e.Controller.DBRepository.UpdateBuildStatus(b.RepositoryHash, b.CommitHash, enums.BuildStatusFinishWithFailure); err != nil {
				b.ErrLogChan <- fmt.Sprintf("GetJobFor %s failed during UpdateBuildStatus err: %s", ev.JobID, err)
			}
			return
		}
		if err := b.StartJob(nextJob, true); err != nil {
			b.ErrLogChan <- err.Error()
			if err := e.Controller.DBRepository.UpdateBuildStatus(b.RepositoryHash, b.CommitHash, enums.BuildStatusFinishWithFailure); err != nil {
				b.ErrLogChan <- fmt.Sprintf("GetJobFor StartJob when ErrorJobFinishReason  %s failed during UpdateBuildStatus err: %s", ev.JobID, err)
			}
		}
		return
	case agent.CompleteJobFinishReason:
		b.LogChan <- ev.Message

		nextJob := b.Triggers.GetJobFor(ev.JobID, true)
		if nextJob != nil {
			if err := b.StartJob(nextJob, true); err != nil {
				b.ErrLogChan <- err.Error()
				if err := e.Controller.DBRepository.UpdateBuildStatus(b.RepositoryHash, b.CommitHash, enums.BuildStatusFinishWithFailure); err != nil {
					b.ErrLogChan <- fmt.Sprintf("GetJobFor StartJob when CompleteJobFinishReason %s failed during UpdateBuildStatus err: %s", ev.JobID, err)
				}
			}
			return
		}
		// Go to looking for another job
	}

	if ev.WasConditionalJob {
		// If finished job was executed in conditional flow
		// we shouldn't never go to normal iterating path
		// It exactly means end of pipeline!
		if err := e.Controller.DBRepository.UpdateBuildStatus(b.RepositoryHash, b.CommitHash, enums.BuildStatusFinishWithSucces); err != nil {
			b.ErrLogChan <- fmt.Sprintf("WasConditionalJob StartJob when ErrorJobFinishReason  %s failed during UpdateBuildStatus err: %s", ev.JobID, err)
		}

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

		if err := b.StartJob(&j, false); err != nil {
			b.ErrLogChan <- err.Error()
			if err := e.Controller.DBRepository.UpdateBuildStatus(b.RepositoryHash, b.CommitHash, enums.BuildStatusFinishWithFailure); err != nil {
				b.ErrLogChan <- fmt.Sprintf("StartJob  %s failed during UpdateBuildStatus err: %s", ev.JobID, err)
			}
		}
		return
	}

	if err := e.Controller.DBRepository.UpdateBuildStatus(b.RepositoryHash, b.CommitHash, enums.BuildStatusFinishWithSucces); err != nil {
		b.ErrLogChan <- fmt.Sprintf("Job %s failed during UpdateBuildStatus - end pipeline err: %s", ev.JobID, err)
	}

}

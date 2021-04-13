package eventhandlers

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/agent/utils"
	"github.com/hetacode/heta-ci/events/agent"
	"github.com/hetacode/heta-ci/events/controller"
	"github.com/hetacode/heta-ci/structs"
)

type StartJobCommandHandler struct {
	App              *app.App
	pipelineTriggers *utils.PipelineTriggers
	buildID          string
}

func (h *StartJobCommandHandler) Handle(event goeh.Event) {
	h.pipelineTriggers = utils.NewPipelineTriggers()
	ev := event.(*controller.StartJobCommand)
	j := ev.Job
	h.sendInfoLog(ev.BuildID, j.ID, fmt.Sprintf("run '%s' job", j.DisplayName))

	h.pipelineTriggers.RegisterTasksTriggers(j)
	h.buildID = ev.BuildID
	// h.pipelineEnvironments.SetCurrent(p.pipeline, &j)

	c := utils.NewContainer(j.Runner, h.App.ArtifactsHostDir, h.App.ArtifactsHostDir)
	defer c.Dispose()

	c.CreateDir(path.Join(utils.JobDir, j.ID))

	var lastFailedTask *structs.Task
	var lastFailedTaskErr error
	for _, t := range j.Tasks {
		// Task with conditions shouldn't be run in normal flow
		if len(t.Conditons) != 0 {
			continue
		}

		if err := h.executeTask(&t, j.ID, c, h.App.ScriptsHostDir); err != nil {
			lastFailedTask = &t
			lastFailedTaskErr = err
			h.sendErrorLog(ev.BuildID, j.ID, err.Error())
			break
		} else {
			if err := h.executeConditionalTask(&t, j.ID, c, h.App.ScriptsHostDir, true); err != nil {
				break
			}
		}
	}

	if lastFailedTask != nil {
		h.executeConditionalTask(lastFailedTask, j.ID, c, h.App.ScriptsHostDir, false)
		// TODO: error should contain error code as separate field
		h.returnError(1, ev.BuildID, j.ID, lastFailedTaskErr.Error())
		return
	}

	h.returnSuccess(ev.BuildID, j.ID, fmt.Sprintf("job '%s' finished", j.DisplayName))
}

func (p *StartJobCommandHandler) executeConditionalTask(t *structs.Task, jobID string, c *utils.Container, scriptsDir string, onSuccess bool) error {
	if t == nil {
		return nil
	}

	conditionalTask := p.pipelineTriggers.GetTaskFor(*t, jobID, onSuccess)
	if conditionalTask == nil {
		return nil
	}

	if err := p.executeTask(conditionalTask, jobID, c, scriptsDir); err != nil {
		p.sendErrorLog(p.buildID, jobID, err.Error())

		p.executeConditionalTask(conditionalTask, jobID, c, scriptsDir, false)
		return err
	} else {
		return p.executeConditionalTask(conditionalTask, jobID, c, scriptsDir, true)
	}
}

func (p *StartJobCommandHandler) executeTask(t *structs.Task, jobID string, c *utils.Container, scriptsDir string) error {
	p.sendInfoLog(p.buildID, jobID, fmt.Sprintf("run '%s' task", t.DisplayName))

	// p.pipelineEnvironments.SetCurrenTask(&t)

	// Prepare script file
	uid, _ := uuid.GenerateUUID()
	filename := uid + ".sh"
	script := createTaskScriptAsBytes(t.Command)
	f, err := os.Create(path.Join(scriptsDir, filename))
	if err != nil {
		return fmt.Errorf("execute task '%s' - create script err: %s", t.DisplayName, err)
	}
	_, err = f.Write(script)
	if err != nil {
		return fmt.Errorf("execute task '%s' - save script err: %s", t.DisplayName, err)

	}
	f.Chmod(775) // execute
	f.Close()

	// Execute script inside container
	// if err := c.ExecuteScript(filename, p.logChannel, p.pipelineEnvironments.GetEnvironments()); err != nil {
	// 	return err
	// }

	p.sendInfoLog(p.buildID, jobID, fmt.Sprintf("task '%s' done", t.DisplayName))
	return nil
}

func createTaskScriptAsBytes(cmd []string) []byte {
	oneCmd := strings.Join(cmd, "\n")
	return []byte(oneCmd)
}

func (h *StartJobCommandHandler) returnSuccess(buildID, jobID, message string) {
	uid, _ := uuid.GenerateUUID()
	ev := &agent.JobFinishedEvent{
		EventData: &goeh.EventData{ID: uid},
		AgentID:   h.App.Config.AgentID,
		BuildID:   buildID,
		JobID:     jobID,
		Reason:    agent.CompleteJobFinishReason,
		ErrorCode: 0,
		Message:   message,
	}
	h.App.MessagingService.SendMessage(ev)
}

func (h *StartJobCommandHandler) returnError(errorCode int, buildID, jobID, message string) {
	uid, _ := uuid.GenerateUUID()
	ev := &agent.JobFinishedEvent{
		EventData: &goeh.EventData{ID: uid},
		AgentID:   h.App.Config.AgentID,
		BuildID:   buildID,
		JobID:     jobID,
		Reason:    agent.ErrorJobFinishReason,
		ErrorCode: errorCode,
		Message:   message,
	}
	h.App.MessagingService.SendMessage(ev)
}

func (h *StartJobCommandHandler) sendInfoLog(buildID, jobID, log string) {
	uid, _ := uuid.GenerateUUID()
	ev := &agent.LogMessageEvent{
		EventData: &goeh.EventData{ID: uid},
		AgentID:   h.App.Config.AgentID,
		BuildID:   buildID,
		JobID:     jobID,
		LogType:   agent.InfoLogType,
		Message:   log,
	}
	h.App.MessagingService.SendMessage(ev)
}

func (h *StartJobCommandHandler) sendErrorLog(buildID, jobID, log string) {
	uid, _ := uuid.GenerateUUID()
	ev := &agent.LogMessageEvent{
		EventData: &goeh.EventData{ID: uid},
		AgentID:   h.App.Config.AgentID,
		BuildID:   buildID,
		JobID:     jobID,
		LogType:   agent.ErrorLogType,
		Message:   log,
	}
	h.App.MessagingService.SendMessage(ev)
}

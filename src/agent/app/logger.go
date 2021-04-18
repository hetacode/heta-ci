package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/hashicorp/go-uuid"
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/commons"
	"github.com/hetacode/heta-ci/events/agent"
)

type Logger struct {
	app *App
}

func NewLogger(a *App) *Logger {
	l := &Logger{
		app: a,
	}

	return l
}

func (h *Logger) ReturnSuccess(buildID, jobID, message string, isConditionalJob bool) {
	createArtifactsPackage(h.app.ArtifactsHostDir, h.app.ArtifactsHostOutDir, buildID, jobID)

	uid, _ := uuid.GenerateUUID()
	ev := &agent.JobFinishedEvent{
		EventData:         &goeh.EventData{ID: uid},
		AgentID:           h.app.Config.AgentID,
		BuildID:           buildID,
		JobID:             jobID,
		Reason:            agent.CompleteJobFinishReason,
		ErrorCode:         0,
		Message:           message,
		WasConditionalJob: isConditionalJob,
	}
	h.app.MessagingService.SendMessage(ev)
	log.Printf("\033[97mfinished job %s\033[0m", jobID)
}

func (h *Logger) ReturnError(errorCode int, buildID, jobID, message string, isConditionalJob bool) {
	createArtifactsPackage(h.app.ArtifactsHostDir, h.app.ArtifactsHostOutDir, buildID, jobID)

	uid, _ := uuid.GenerateUUID()
	ev := &agent.JobFinishedEvent{
		EventData:         &goeh.EventData{ID: uid},
		AgentID:           h.app.Config.AgentID,
		BuildID:           buildID,
		JobID:             jobID,
		Reason:            agent.ErrorJobFinishReason,
		ErrorCode:         errorCode,
		Message:           message,
		WasConditionalJob: isConditionalJob,
	}
	h.app.MessagingService.SendMessage(ev)
	log.Printf("\033[31mfinished job %s with error\033[0m", jobID)
}

func (h *Logger) SendInfoLog(buildID, jobID, log string) {
	uid, _ := uuid.GenerateUUID()
	ev := &agent.LogMessageEvent{
		EventData: &goeh.EventData{ID: uid},
		AgentID:   h.app.Config.AgentID,
		BuildID:   buildID,
		JobID:     jobID,
		LogType:   agent.InfoLogType,
		Message:   log,
	}
	h.app.MessagingService.SendMessage(ev)
}

func (h *Logger) SendErrorLog(buildID, jobID, log string) {
	uid, _ := uuid.GenerateUUID()
	ev := &agent.LogMessageEvent{
		EventData: &goeh.EventData{ID: uid},
		AgentID:   h.app.Config.AgentID,
		BuildID:   buildID,
		JobID:     jobID,
		LogType:   agent.ErrorLogType,
		Message:   log,
	}
	h.app.MessagingService.SendMessage(ev)
}

func createArtifactsPackage(artifactsDirPath, artifatcsOutDirPath, buildID, jobID string) {
	b, err := commons.ArchiveDirectory(artifatcsOutDirPath)
	if err != nil {
		log.Printf("zip err: %s", err)
	} else {
		ioutil.WriteFile(path.Join(artifactsDirPath, fmt.Sprintf("artifacts_%s_%s.zip", buildID, jobID)), b, 0644)
	}
}

package eventhandlers

import (
	"fmt"
	"log"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/agent/executors"
	"github.com/hetacode/heta-ci/agent/utils"
	"github.com/hetacode/heta-ci/events/controller"
)

type StartJobCommandHandler struct {
	App                  *app.App
	pipelineEnvironments *utils.PipelineEnvironments
	pipelineTriggers     *utils.PipelineTriggers
	buildID              string
}

func (h *StartJobCommandHandler) Handle(event goeh.Event) {
	logger := app.NewLogger(h.App)
	h.pipelineTriggers = utils.NewPipelineTriggers()
	h.pipelineEnvironments = utils.NewPipelineEnvironments(h.App.ScriptsHostDir, h.App.ArtifactsHostDir)

	ev := event.(*controller.StartJobCommand)
	j := ev.Job

	h.pipelineTriggers.RegisterTasksTriggers(j)
	h.buildID = ev.BuildID

	if ev.HasArtifacts {
		fileBytes, err := h.App.ArtifactsService.DownloadArtifacts(h.buildID)
		if err != nil {
			logger.ReturnError(1, h.buildID, j.ID, fmt.Sprintf("start job | download artifacts failed | err: %s", err), ev.IsConditional)
			return
		}

		log.Fatalf("unimplemented %s", fileBytes)
	}

	je := executors.NewJobExecutor(h.App, logger, h.pipelineEnvironments, h.pipelineTriggers, &j, ev.PipelineID, ev.BuildID, ev.IsConditional)
	je.Execute()
}

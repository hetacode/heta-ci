package eventhandlers

import (
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

	je := executors.NewJobExecutor(h.App, logger, h.pipelineEnvironments, h.pipelineTriggers, &j, ev.PipelineID, ev.BuildID, ev.IsConditional, ev.HasArtifacts)
	je.Execute()
}

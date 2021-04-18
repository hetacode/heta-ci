package executors

import (
	"fmt"
	"os"

	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/agent/utils"
	"github.com/hetacode/heta-ci/structs"
)

type JobExecutor struct {
	app                  *app.App
	logger               *app.Logger
	pipelineEnvironments *utils.PipelineEnvironments
	pipelineTriggers     *utils.PipelineTriggers
	pipelineID           string
	buildID              string
	job                  *structs.Job
	isConditional        bool
}

func NewJobExecutor(
	a *app.App,
	logger *app.Logger,
	pipelineEnvironments *utils.PipelineEnvironments,
	pipelineTriggers *utils.PipelineTriggers,
	job *structs.Job,
	pipelineID string,
	buildID string,
	isConditional bool,
) *JobExecutor {
	e := &JobExecutor{
		app:                  a,
		logger:               logger,
		job:                  job,
		pipelineID:           pipelineID,
		buildID:              buildID,
		isConditional:        isConditional,
		pipelineTriggers:     utils.NewPipelineTriggers(),
		pipelineEnvironments: utils.NewPipelineEnvironments(a.ScriptsHostDir, a.ArtifactsHostDir),
	}
	e.logger = app.NewLogger(e.app)

	return e
}

func (e *JobExecutor) Execute() {
	e.logger.SendInfoLog(e.buildID, e.job.ID, fmt.Sprintf("run '%s' job", e.job.DisplayName))

	os.RemoveAll(e.app.ScriptsHostDir)
	os.RemoveAll(e.app.ArtifactsHostDir)

	if err := e.createInitDirs(); err != nil {
		e.logger.ReturnError(1, e.buildID, e.job.ID, err.Error(), e.isConditional)
		return
	}
	e.pipelineEnvironments.SetCurrent(e.pipelineID, e.job.DisplayName)

	c := utils.NewContainer(e.job.Runner, e.app.ScriptsHostDir, e.app.ArtifactsHostDir)
	defer c.Dispose()

	c.CreateDir(utils.TasksDir)

	te := NewTasksExecutor(e.app, e.pipelineEnvironments, e.pipelineTriggers, c, e.logger, e.job.Tasks, e.buildID, e.job.ID, e.app.ScriptsHostDir, e.isConditional)
	if err := te.Execute(); err == nil {
		e.logger.ReturnSuccess(e.buildID, e.job.ID, fmt.Sprintf("job '%s' finished", e.job.DisplayName), e.isConditional)
	}
}

func (e *JobExecutor) createInitDirs() error {
	if err := os.Mkdir(e.app.ScriptsHostDir, os.ModePerm); err != nil {

		return fmt.Errorf("create scripts temp directory err: %s", err)
	}

	if err := os.MkdirAll(e.app.ArtifactsHostInDir, os.ModePerm); err != nil {
		return fmt.Errorf("create artifacts in temp directory err: %s", err)
	}

	if err := os.MkdirAll(e.app.ArtifactsHostOutDir, os.ModePerm); err != nil {
		return fmt.Errorf("create artifacts out temp directory err: %s", err)
	}

	return nil
}

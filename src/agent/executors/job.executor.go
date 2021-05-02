package executors

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/agent/errors"
	"github.com/hetacode/heta-ci/agent/utils"
	"github.com/hetacode/heta-ci/commons"
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
	hasArtificats        bool
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
	hasArtificats bool,
) *JobExecutor {
	e := &JobExecutor{
		app:                  a,
		logger:               logger,
		job:                  job,
		pipelineID:           pipelineID,
		buildID:              buildID,
		isConditional:        isConditional,
		hasArtificats:        hasArtificats,
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

	if err := e.downloadAndExtractArtifactsPackage(); err != nil {
		e.logger.ReturnError(1, e.buildID, e.job.ID, err.Error(), e.isConditional)
		return
	}

	c := utils.NewContainer(e.job.Runner, e.app.ScriptsHostDir, e.app.ArtifactsHostDir)
	defer c.Dispose()

	c.CreateDir(utils.TasksDir)

	te := NewTasksExecutor(e.app, e.pipelineEnvironments, e.pipelineTriggers, c, e.logger, e.job.Tasks, e.buildID, e.job.ID, e.app.ScriptsHostDir, e.isConditional)
	if err := te.Execute(); err == nil {
		createArtifactsPackage(e.app.ArtifactsHostDir, e.app.ArtifactsHostOutDir, e.buildID, e.job.ID)
		if err := e.uploadArtifacts(); err != nil {
			e.logger.ReturnError(1, e.buildID, e.job.ID, err.Error(), e.isConditional)
		} else {
			e.logger.ReturnSuccess(e.buildID, e.job.ID, fmt.Sprintf("job '%s' finished", e.job.DisplayName), e.isConditional)
		}
	} else if te, ok := err.(*errors.ContainerError); ok {
		e.logger.ReturnError(te.ErrorCode, e.buildID, e.job.ID, err.Error(), e.isConditional)
	} else {
		e.logger.ReturnError(1, e.buildID, e.job.ID, err.Error(), e.isConditional)
	}

}

func (j *JobExecutor) downloadAndExtractArtifactsPackage() error {
	if j.hasArtificats {
		fileBytes, err := j.app.ArtifactsService.DownloadArtifacts(j.buildID)
		if err != nil {
			return fmt.Errorf("start job | download artifacts failed | err: %s", err)
		}
		if err := commons.ExtractDirectory(fileBytes, j.app.ArtifactsHostInDir); err != nil {
			return fmt.Errorf("start job | extract artifacts file failed | err: %s", err)
		}
	}

	return nil
}

func createArtifactsPackage(artifactsDirPath, artifatcsOutDirPath, buildID, jobID string) {
	b, err := commons.ArchiveDirectory(artifatcsOutDirPath)
	if err != nil {
		log.Printf("zip err: %s", err)
	} else {
		ioutil.WriteFile(path.Join(artifactsDirPath, utils.ArtifactsFileName(buildID, jobID)), b, 0644)
	}
}

func (e *JobExecutor) uploadArtifacts() error {
	exists, err := commons.IsFileExists(path.Join(e.app.ArtifactsHostDir, utils.ArtifactsFileName(e.buildID, e.job.ID)))
	if exists {
		bytes, err := ioutil.ReadFile(path.Join(e.app.ArtifactsHostDir, utils.ArtifactsFileName(e.buildID, e.job.ID)))
		if err != nil {
			return fmt.Errorf("couldn't read artifacts file | err: %s", err)
		}

		e.app.ArtifactsService.UploadArtifacts(e.buildID, e.job.ID, bytes)
	}
	if err != nil {
		return fmt.Errorf("cannot upload artifacts | err: %s", err)
	}

	return nil
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

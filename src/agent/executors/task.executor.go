package executors

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/agent/app"
	"github.com/hetacode/heta-ci/agent/utils"
	"github.com/hetacode/heta-ci/structs"
)

type TasksExecutor struct {
	app                  *app.App
	tasks                []structs.Task
	container            *utils.Container
	pipelineEnvironments *utils.PipelineEnvironments
	pipelineTriggers     *utils.PipelineTriggers
	buildID              string
	jobID                string
	scriptsDir           string
	isConditionalJob     bool
	logger               *app.Logger
}

func NewTasksExecutor(
	app *app.App,
	pipelineEnvironments *utils.PipelineEnvironments,
	pipelineTriggers *utils.PipelineTriggers,
	container *utils.Container,
	logger *app.Logger,
	tasks []structs.Task, buildID, jobID, scriptsDir string, isConditionalJob bool) *TasksExecutor {
	t := &TasksExecutor{
		app:                  app,
		pipelineEnvironments: pipelineEnvironments,
		pipelineTriggers:     pipelineTriggers,
		container:            container,
		logger:               logger,
		isConditionalJob:     isConditionalJob,
		tasks:                tasks,
		buildID:              buildID,
		jobID:                jobID,
		scriptsDir:           scriptsDir,
	}
	return t
}

func (t *TasksExecutor) Execute() error {
	var lastFailedTask *structs.Task
	var lastFailedTaskErr error
	for _, task := range t.tasks {
		if len(task.Conditons) != 0 {
			return nil
		}

		if err := t.executeTask(&task, t.buildID, t.jobID, t.container, t.app.ScriptsHostDir); err != nil {
			lastFailedTask = &task
			lastFailedTaskErr = err
			t.logger.SendErrorLog(t.buildID, t.jobID, err.Error())
			break
		} else {
			if err := t.executeConditionalTask(&task, t.buildID, t.jobID, t.container, t.app.ScriptsHostDir, true); err != nil {
				break
			}
		}
	}

	if lastFailedTask != nil {
		t.executeConditionalTask(lastFailedTask, t.buildID, t.jobID, t.container, t.app.ScriptsHostDir, false)

		return lastFailedTaskErr
	}

	return nil
}

func (p *TasksExecutor) executeConditionalTask(t *structs.Task, buildID, jobID string, c *utils.Container, scriptsDir string, onSuccess bool) error {
	if t == nil {
		return nil
	}

	conditionalTask := p.pipelineTriggers.GetTaskFor(*t, jobID, onSuccess)
	if conditionalTask == nil {
		return nil
	}

	if err := p.executeTask(conditionalTask, buildID, jobID, c, scriptsDir); err != nil {
		p.logger.SendErrorLog(p.buildID, jobID, err.Error())

		p.executeConditionalTask(conditionalTask, buildID, jobID, c, scriptsDir, false)
		return err
	} else {
		return p.executeConditionalTask(conditionalTask, buildID, jobID, c, scriptsDir, true)
	}
}

func (p *TasksExecutor) executeTask(t *structs.Task, buildID, jobID string, c *utils.Container, scriptsDir string) error {
	p.logger.SendInfoLog(buildID, jobID, fmt.Sprintf("run '%s' task", t.DisplayName))

	p.pipelineEnvironments.SetCurrenTask(t)

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
	msg, err := c.ExecuteScript(filename, p.pipelineEnvironments.GetEnvironments())
	p.logger.SendInfoLog(buildID, jobID, msg)
	if err != nil {
		return err
	}

	p.logger.SendInfoLog(buildID, jobID, fmt.Sprintf("task '%s' done", t.DisplayName))
	return nil
}

func createTaskScriptAsBytes(cmd []string) []byte {
	oneCmd := strings.Join(cmd, "\n")
	return []byte(oneCmd)
}

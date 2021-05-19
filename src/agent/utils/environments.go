package utils

import (
	"fmt"
	"path"

	"github.com/hetacode/heta-ci/structs"
)

const (
	AgentCodeDirEnvName            = "AGENT_CODE_DIR"              // directory where exists downloaded source code of build pipeline
	AgentScriptsDirEnvName         = "AGENT_SCRIPTS_DIR"           // directory where lands tasks commands scripts
	AgentJobArtifactsInDirEnvName  = "AGENT_JOB_ARTIFACTS_IN_DIR"  // directory where land downloaded artifacts
	AgentJobArtifactsOutDirEnvName = "AGENT_JOB_ARTIFACTS_OUT_DIR" // directory where task can put files that will were upload to the controller at the end of job

	AgentTasksDirEnvName   = "AGENT_TASKS_DIR"
	AgentJobIDEnvName      = "AGENT_JOB_ID"
	AgentJobNameEnvName    = "AGENT_JOB_NAME"
	AgentPipelineIDEnvName = "AGENT_PIPELINE_ID" // should be create by controller as unique uuid
	AgentTaskIDEnvName     = "AGENT_TASK_ID"
	AgentTaskNameEnvName   = "AGENT_TASK_NAME"
)

type PipelineEnvironments struct {
	Env map[string]string
}

func NewPipelineEnvironments(scriptsDir, jobArtifactsDir string) *PipelineEnvironments {
	p := &PipelineEnvironments{
		Env: map[string]string{
			AgentCodeDirEnvName:            ContainerCodeDir,
			AgentScriptsDirEnvName:         ContainerScriptsDir,
			AgentJobArtifactsInDirEnvName:  path.Join(ContainerArtifactsDir, "in"),
			AgentJobArtifactsOutDirEnvName: path.Join(ContainerArtifactsDir, "out"),
			AgentTasksDirEnvName:           "/tasks",
		},
	}
	return p
}

func (p *PipelineEnvironments) SetCurrent(piplineID, jobName string) {
	p.Env[AgentJobIDEnvName] = piplineID
	p.Env[AgentJobNameEnvName] = jobName
}

func (p *PipelineEnvironments) SetCurrenTask(t *structs.Task) {
	p.Env[AgentTaskIDEnvName] = t.ID
	p.Env[AgentTaskNameEnvName] = t.DisplayName
}

func (p *PipelineEnvironments) GetAllEnvNames() []string {
	envs := make([]string, 5)
	for k := range p.Env {
		envs = append(envs, k)
	}

	return envs
}

func (p *PipelineEnvironments) GetEnvironments() []string {
	env := make([]string, len(p.Env))

	counter := 0
	for k, v := range p.Env {
		env[counter] = fmt.Sprintf("%s=%s", k, v)
		counter++
	}
	return env
}

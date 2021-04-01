package main

import (
	"fmt"

	"github.com/hetacode/heta-ci/structs"
)

const (
	AgentScriptsDirEnvName           = "AGENT_SCRIPTS_DIR"
	AgentJobArtifactsDirEnvName      = "AGENT_JOB_ARTIFACTS_DIR"
	AgentPipelineArtifactsDirEnvName = "AGENT_PIPELINE_ARTIFACTS_DIR"
	AgentJobIDEnvName                = "AGENT_JOB_ID"
	AgentJobNameEnvName              = "AGENT_JOB_NAME"
	AgentPipelineIDEnvName           = "AGENT_PIPELINE_ID" // should be create by controller as unique uuid
	AgentPipelineNameEnvName         = "AGENT_PIPELINE_NAME"
	AgentTaskIDEnvName               = "AGENT_TASK_ID"
	AgentTaskNameEnvName             = "AGENT_TASK_NAME"
)

type PipelineEnvironments struct {
	Env map[string]string
}

func NewPipelineEnvironments(scriptsDir, jobsDir, pipelineDir string) *PipelineEnvironments {
	p := &PipelineEnvironments{
		Env: map[string]string{
			AgentScriptsDirEnvName:           scriptsDir,
			AgentJobArtifactsDirEnvName:      jobsDir,
			AgentPipelineArtifactsDirEnvName: pipelineDir,
		},
	}
	return p
}

func (p *PipelineEnvironments) SetCurrent(pi *structs.Pipeline, j *structs.Job) {
	p.Env[AgentPipelineNameEnvName] = pi.Name
	p.Env[AgentJobIDEnvName] = j.ID
	p.Env[AgentJobNameEnvName] = j.DisplayName
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

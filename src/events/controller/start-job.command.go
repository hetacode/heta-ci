package controller

import goeh "github.com/hetacode/go-eh"

// StartJobCommand just assign pipeline job to the agent
// Full job data - like  code, job steps - should downloaded via rest api
type StartJobCommand struct {
	*goeh.EventData
	BuildID    string `json:"build_id"`
	PipelineID string `json:"pipeline_id"`
	JobID      string `json:"job_id"`
}

func (e *StartJobCommand) GetType() string {
	return "StartJobCommand"
}

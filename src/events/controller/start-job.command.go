package controller

import (
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/structs"
)

// StartJobCommand just assign pipeline job to the agent
// Full job data - like  code, job steps - should downloaded via rest api
type StartJobCommand struct {
	*goeh.EventData
	BuildID    string      `json:"build_id"`
	PipelineID string      `json:"pipeline_id"`
	Job        structs.Job `json:"job"`
}

func (e *StartJobCommand) GetType() string {
	return "StartJobCommand"
}

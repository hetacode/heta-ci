package utils

import (
	"github.com/hetacode/heta-ci/structs"
)

type PipelineTriggers struct {
	Triggers map[string][]ToExecute
}

type ToExecute struct {
	ConditionType structs.ConditionType
	Job           *structs.Job
}

func NewPipelineTriggers() *PipelineTriggers {
	t := &PipelineTriggers{
		Triggers: make(map[string][]ToExecute),
	}

	return t
}

// RegisterJob in conditions resolver
func (t *PipelineTriggers) RegisterJobsFor(pipeline *structs.Pipeline) {
	for _, job := range pipeline.Jobs {
		if len(job.Conditons) == 0 {
			continue
		}

		for _, c := range job.Conditons {
			triggers := t.Triggers[c.On]
			triggers = append(triggers, ToExecute{
				ConditionType: c.Type,
				Job:           &job,
			})
			t.Triggers[c.On] = triggers
		}
	}
}

// GetJobFor for given jobID.
// Returned job should be fitted to condition for finished with success or not of previous job
func (t *PipelineTriggers) GetJobFor(jobID string, isSuccess bool) *structs.Job {
	triggers := t.Triggers[jobID]
	for _, tr := range triggers {
		if isSuccess {
			switch tr.ConditionType {
			case structs.OnSuccess:
				return tr.Job
			}
		} else {
			switch tr.ConditionType {
			case structs.OnFailure:
				return tr.Job
			}
		}
	}

	return nil
}

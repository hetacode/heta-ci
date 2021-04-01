package main

import (
	"fmt"

	"github.com/hetacode/heta-ci/structs"
)

type PipelineTriggers struct {
	// Key should jobID or pair jobID|taskID
	Triggers map[string][]ToExecute
}

type ToExecute struct {
	ConditionType structs.ConditionType
	Job           *structs.Job
	Task          *structs.Task
}

func NewPipelineTriggers() *PipelineTriggers {
	t := &PipelineTriggers{
		Triggers: make(map[string][]ToExecute),
	}

	return t
}

func (t *PipelineTriggers) RegisterJob(job structs.Job) {
	if len(job.Conditons) == 0 {
		return
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

func (t *PipelineTriggers) RegisterTask(jobID string, task structs.Task) {
	if len(task.Conditons) == 0 {
		return
	}

	for _, c := range task.Conditons {
		key := fmt.Sprintf("%s|%s", jobID, c.On)
		triggers := t.Triggers[key]
		triggers = append(triggers, ToExecute{
			ConditionType: c.Type,
			Task:          &task,
		})
		t.Triggers[key] = triggers
	}
}

func (t *PipelineTriggers) GetJobFor(job structs.Job, isSuccess bool) *structs.Job {
	triggers := t.Triggers[job.ID]
	for _, t := range triggers {
		if isSuccess {
			switch t.ConditionType {
			case structs.OnSuccess:
				return t.Job
			}
		} else {
			switch t.ConditionType {
			case structs.OnFailure:
				return t.Job
			}
		}
	}

	return nil
}

func (t *PipelineTriggers) GetTaskFor(task structs.Task, jobID string, isSuccess bool) *structs.Task {
	key := fmt.Sprintf("%s|%s", jobID, task.ID)
	triggers := t.Triggers[key]
	for _, t := range triggers {
		if isSuccess {
			switch t.ConditionType {
			case structs.OnSuccess:
				return t.Task
			}
		} else {
			switch t.ConditionType {
			case structs.OnFailure:
				return t.Task
			}
		}
	}

	return nil
}

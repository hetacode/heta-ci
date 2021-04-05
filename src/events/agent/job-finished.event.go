package agent

import goeh "github.com/hetacode/go-eh"

type JobFinishReason string

const (
	CompleteJobFinishReason JobFinishReason = "complete"
	ErrorJobFinishReason                    = "error"
)

type JobFinishedEvent struct {
	*goeh.EventData
	AgentID   string          `json:"agent_id"`
	Reason    JobFinishReason `json:"reason"`
	BuildID   string          `json:"build_id"`
	JobID     string          `json:"job_id"`
	Message   string          `json:"message"`
	ErrorCode int             `json:"err_code"`
}

func (e *JobFinishedEvent) GetType() string {
	return "JobFinishedEvent"
}

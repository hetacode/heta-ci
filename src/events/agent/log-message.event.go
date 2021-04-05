package agent

import goeh "github.com/hetacode/go-eh"

type LogType string

const (
	InfoLogType  LogType = "info_log"
	ErrorLogType         = "error_log"
)

type LogMessageEvent struct {
	*goeh.EventData
	AgentID string  `json:"agent_id"`
	Type    LogType `json:"log_type"`
	BuildID string  `json:"build_id"`
	JobID   string  `json:"job_id"`
	Message string  `json:"message"`
}

func (e *LogMessageEvent) GetType() string {
	return "LogMessageEvent"
}

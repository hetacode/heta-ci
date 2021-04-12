package controller

import goeh "github.com/hetacode/go-eh"

// AgentConfirmedEvent is send just after connect agent to controller in order to
// confirm connection and send generated id of agent
type AgentConfirmedEvent struct {
	*goeh.EventData
	// AgentID is unique id generated just after connect agent to the controller
	AgentID string `json:"agent_id"`
}

func (e *AgentConfirmedEvent) GetType() string {
	return "AgentConfirmedEvent"
}

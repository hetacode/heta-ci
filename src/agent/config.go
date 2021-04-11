package main

import (
	"log"
	"os"
)

type Config struct {
	AgentID  string
	Hostname string
}

func NewConfig(agentID string) *Config {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("agent | config get hostaname | err: %s", err)
	}
	c := &Config{
		Hostname: hostname,
		AgentID:  agentID,
	}

	return c
}

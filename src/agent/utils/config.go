package utils

import (
	"fmt"
	"log"
	"os"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/events"
)

type Config struct {
	AgentID      string
	Hostname     string
	EventsMapper *goeh.EventsMapper
}

func NewConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("agent | config get hostaname | err: %s", err)
	}
	c := &Config{
		Hostname:     hostname,
		EventsMapper: events.NewEventsMapper(),
	}

	return c
}

func ArtifactsFileName(buildID, jobID string) string {
	return fmt.Sprintf("artifacts_%s_%s.zip", buildID, jobID)
}

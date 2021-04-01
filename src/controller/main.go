package main

import (
	"fmt"

	"github.com/hetacode/heta-ci/structs"
)

func main() {
	c := NewController()
	c.AddPipeline(preparePipeline())

	fmt.Print("controller")
}

func preparePipeline() *structs.Pipeline {
	pipeline := &structs.Pipeline{
		Name: "Test shell scripts in one container",
		Jobs: []structs.Job{
			{
				ID:          "test_alpine",
				DisplayName: "Alpine runner",
				Runner:      "alpine",
				Tasks: []structs.Task{
					{
						ID:          "correct",
						DisplayName: "Correct script",
						Command: []string{
							"echo Start",
							"cd /etc && ls -al",
							"echo End",
						},
					},
					{
						ID:          "correct",
						DisplayName: "Correct script",
						Command: []string{
							"echo Start",
							"echo job artifacts dir: $AGENT_JOB_ARTIFACTS_DIR",
							"echo scripts dir: $AGENT_SCRIPTS_DIR",
							"echo End",
						},
					},
				},
			},
		},
	}

	return pipeline
}

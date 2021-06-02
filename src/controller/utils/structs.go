package utils

import (
	"fmt"

	"github.com/hetacode/heta-ci/structs"
)

type BuildLastCommits map[string]string

func (b BuildLastCommits) Add(repositoryID string, runOnType structs.RunOnType, triggerValue, commitHash string) {
	key := fmt.Sprintf("%s-%s-%s", repositoryID, runOnType, triggerValue)
	b[key] = commitHash
}

func (b BuildLastCommits) Get(repositoryID string, runOnType structs.RunOnType, triggerValue string) *string {
	key := fmt.Sprintf("%s-%s-%s", repositoryID, runOnType, triggerValue)
	v, ok := b[key]
	if !ok {
		return nil
	}
	return &v
}

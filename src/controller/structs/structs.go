package utils

import (
	"fmt"

	"github.com/hetacode/heta-ci/controller/db"
	"github.com/hetacode/heta-ci/structs"
)

type BuildLastCommits map[string]string

func (b BuildLastCommits) Add(dbRepository db.DBRepository, repositoryID string, runOnType structs.RunOnType, triggerValue, commitHash string) {
	key := fmt.Sprintf("%s-%s-%s", repositoryID, runOnType, triggerValue)
	b[key] = commitHash
	go dbRepository.SetLastBuildCommit(key, commitHash)
}

func (b BuildLastCommits) Get(dbRepository db.DBRepository, repositoryID string, runOnType structs.RunOnType, triggerValue string) *string {
	key := fmt.Sprintf("%s-%s-%s", repositoryID, runOnType, triggerValue)
	v, ok := b[key]
	if !ok {
		value, _, _ := dbRepository.GetLastBuildCommit(key)
		if value != nil {
			b[key] = *value
			return value
		}
		return nil
	}
	return &v
}

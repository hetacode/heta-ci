package processors

import (
	"fmt"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/controller/utils"
)

type RepositoryProcessor struct {
}

func (p *RepositoryProcessor) Process(repo utils.Repository) error {
	// TODO:
	// 1. checkout last changes
	// 2. archive repo

	uid, _ := uuid.GenerateUUID()
	_, err := git.PlainClone(path.Join("repos", uid), false, &git.CloneOptions{
		URL:        repo.Url,
		RemoteName: repo.DefaultBranch,
	})

	if err != nil {
		return fmt.Errorf("clone repo failed %s", err)
	}

	return nil
}

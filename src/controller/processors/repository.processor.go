package processors

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/commons"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/hetacode/heta-ci/structs"
	"gopkg.in/yaml.v2"
)

type RepositoryProcessor struct {
	Controller *utils.Controller
}

func (p *RepositoryProcessor) Process(repo utils.Repository) error {
	// TODO:
	// 2. archive repo

	processingID, _ := uuid.GenerateUUID()
	repoPath := path.Join("repos", processingID)
	if err := cloneRepo(&repo, repoPath); err != nil {
		return err
	}
	defer os.RemoveAll(repoPath)

	repoBytes, err := commons.ArchiveDirectory(repoPath)
	if err != nil {
		return fmt.Errorf("archive repo failed %s", err)
	}

	os.WriteFile(fmt.Sprintf("repos/%s.zip", processingID), repoBytes, 0777)

	pipeline, err := createPipeline(repoPath)
	if err != nil {
		return err
	}
	p.Controller.AddPipeline(pipeline)

	return nil
}

func createPipeline(repoPath string) (*structs.Pipeline, error) {
	b, err := os.ReadFile(path.Join(repoPath, ".heta-ci/pipeline.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read pipeline config file failed %s", err)
	}

	var pipeline *structs.Pipeline
	if err := yaml.Unmarshal(b, &pipeline); err != nil {
		log.Fatal(err)

		return nil, fmt.Errorf("unmarshal pipeline config file failed %s", err)
	}

	return pipeline, nil
}

func cloneRepo(repo *utils.Repository, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:        repo.Url,
		RemoteName: repo.DefaultBranch,
	})

	if err != nil {
		return fmt.Errorf("clone repo failed %s", err)
	}

	return nil
}

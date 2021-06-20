package db

import (
	"github.com/hetacode/heta-ci/controller/enums"
	"github.com/hetacode/heta-ci/structs"
)

type DBRepository interface {
	GetRepositories() (*[]Repository, error)
	GetRepositoriesByProjectID(projectID string) (*[]Repository, error)
	StoreBuildData(id string, pipeline *structs.Pipeline, repositoryHash, commitHash string) error
	UpdateBuildStatus(repositoryHash, commit string, status enums.BuildStatus) error
	GetBuildsByRepositoryHash(repositoryHash string) (*[]Build, error)
	GetBuildBy(repositoryHash, commitHash string) (*Build, error)
	SetLastBuildCommit(key string, commitHash string) error
	GetLastBuildCommit(key string) (commitHash *string, createOn *int64, error error)
}

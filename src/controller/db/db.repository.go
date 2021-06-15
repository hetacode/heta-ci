package db

import "github.com/hetacode/heta-ci/controller/utils"

type DBRepository interface {
	GetRepositories() (*[]Repository, error)
	GetRepositoriesByProjectID(projectID string) (*[]Repository, error)
	StoreBuildData(buildPipeline *utils.PipelineBuild, repositoryHash, commitHash string) error
	GetBuildsByRepositoryHash(repositoryHash string) (*[]Build, error)
	GetBuildByCommitHash(commitHash string) (*Build, error)
	SetLastBuildCommit(key string, commitHash string) error
	GetLastBuildCommit(key string) (commitHash string, createOn int64, error error)
}

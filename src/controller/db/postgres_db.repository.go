package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hetacode/heta-ci/controller/enums"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/xo/dburl"
)

type PostgresDBRepository struct {
	connection *sql.DB
}

func NewPostgresDBRepository(connectionString string) (*PostgresDBRepository, error) {
	d := &PostgresDBRepository{}
	conn, err := dburl.Open(connectionString)
	if err != nil {
		return nil, fmt.Errorf("open connection to database failed %s", err)
	}
	d.connection = conn
	return d, nil
}

func (d *PostgresDBRepository) GetRepositories() (*[]Repository, error) {
	rows, err := d.connection.Query("select * from public.repository")
	if err != nil {
		return nil, fmt.Errorf("cannot query for GetRepositories %s", err)
	}
	return getRepositories(rows), nil
}
func (d *PostgresDBRepository) GetRepositoriesByProjectID(projectID string) (*[]Repository, error) {
	rows, err := d.connection.Query("select * from public.repository where project_uid=?", projectID)
	if err != nil {
		return nil, fmt.Errorf("cannot query for GetRepositoriesByProjectID %s", err)
	}

	return getRepositories(rows), nil
}

func (d *PostgresDBRepository) StoreBuildData(buildPipeline *utils.PipelineBuild, repositoryHash, commitHash string) error {
	pipelineBytes, _ := json.Marshal(buildPipeline.Pipeline)
	uid, _ := uuid.FromString(buildPipeline.ID)
	dbBuild := &Build{
		UID:            uid,
		RepositoryHash: repositoryHash,
		CommitHash:     commitHash,
		PipelineJSON:   string(pipelineBytes),
		ResultStatus:   string(enums.BuildStatusNone),
		CreatedAt:      time.Now().Unix(),
	}
	return dbBuild.Insert(d.connection)
}

func (d *PostgresDBRepository) GetBuildsByRepositoryHash(repositoryHash string) (*[]Build, error) {
	rows, err := d.connection.Query("select * from public.build where repository_hash=?", repositoryHash)
	if err != nil {
		return nil, fmt.Errorf("cannot query for GetBuildsByRepositoryHash %s", err)
	}

	builds := make([]Build, 0)
	for rows.Next() {
		build := Build{}
		err := rows.Scan(&build.UID,
			&build.ResultStatus,
			&build.RepositoryHash,
			&build.PipelineJSON,
			&build.Logs,
			&build.FinishAt,
			&build.CreatedAt,
			&build.CommitHash,
			&build.ArtifactsUID,
		)
		if err != nil {
			return nil, fmt.Errorf("GetBuildsByRepositoryHash scan Build object failed %s", err)
		}
		builds = append(builds, build)
	}

	return &builds, nil
}
func (d *PostgresDBRepository) GetBuildByCommitHash(commitHash string) (*Build, error) {
	row := d.connection.QueryRow("select * from public.build where commit_hash=?", commitHash)

	build := Build{}
	err := row.Scan(&build.UID,
		&build.ResultStatus,
		&build.RepositoryHash,
		&build.PipelineJSON,
		&build.Logs,
		&build.FinishAt,
		&build.CreatedAt,
		&build.CommitHash,
		&build.ArtifactsUID,
	)
	if err != nil {
		return nil, fmt.Errorf("GetBuildByCommitHash scan Build object failed %s", err)
	}

	return &build, nil

}
func (d *PostgresDBRepository) SetLastBuildCommit(key string, commitHash string) error {
	lastCommit := &KvBuildLastCommit{
		Key:             key,
		ValueHashCommit: commitHash,
		CreatedAt:       time.Now().Unix(),
	}
	return lastCommit.Insert(d.connection)

}
func (d *PostgresDBRepository) GetLastBuildCommit(key string) (commitHash *string, createOn *int64, error error) {
	row := d.connection.QueryRow("select * from public.kv_build_last_commit where key=? order by id desc", commitHash)

	lastCommit := KvBuildLastCommit{}
	err := row.Scan(&lastCommit.ID,
		&lastCommit.Key,
		&lastCommit.ValueHashCommit,
		&lastCommit.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("GetLastBuildCommit scan KvBuildLastCommit object failed %s", err)
	}

	return &lastCommit.ValueHashCommit, &lastCommit.CreatedAt, nil
}

func getRepositories(rows *sql.Rows) *[]Repository {
	repositories := make([]Repository, 0)
	for rows.Next() {
		repo := Repository{}
		rows.Scan(&repo.RepoHash, &repo.RepositoryURL, &repo.DefaultBranch, &repo.Name, &repo.CreatedAt, &repo.ProjectUID)
		repositories = append(repositories, repo)
	}

	return &repositories
}

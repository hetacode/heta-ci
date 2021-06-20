package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hetacode/heta-ci/controller/enums"
	"github.com/hetacode/heta-ci/structs"
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
	rows, err := d.connection.Query("select * from public.repository where project_uid=$1", projectID)
	if err != nil {
		return nil, fmt.Errorf("cannot query for GetRepositoriesByProjectID %s", err)
	}

	return getRepositories(rows), nil
}

func (d *PostgresDBRepository) StoreBuildData(id string, pipeline *structs.Pipeline, repositoryHash, commitHash string) error {
	pipelineBytes, _ := json.Marshal(pipeline)
	uid, _ := uuid.FromString(id)
	dbBuild := &Build{
		UID:            uid,
		RepositoryHash: repositoryHash,
		CommitHash:     commitHash,
		PipelineJSON:   string(pipelineBytes),
		Status:         string(enums.BuildStatusNone),
		CreatedAt:      time.Now().Unix(),
	}
	err := dbBuild.Insert(d.connection)
	if err != sql.ErrNoRows {
		return err
	}

	return nil
}

func (d *PostgresDBRepository) UpdateBuildStatus(repositoryHash, commit string, status enums.BuildStatus) error {
	build, err := d.GetBuildBy(repositoryHash, commit)
	if err != nil {
		return fmt.Errorf("UpdateBuildStatus get build failed: %s", err)
	}
	build.Status = string(status)

	if err := build.Update(d.connection); err != nil {
		return fmt.Errorf("update UpdateBuildStatus failed for repo: %s commit: %s %+v| err: %s", repositoryHash, commit, build, err)
	}
	return nil
}

func (d *PostgresDBRepository) GetBuildsByRepositoryHash(repositoryHash string) (*[]Build, error) {
	rows, err := d.connection.Query("select * from public.build where repository_hash=$1", repositoryHash)
	if err != nil {
		return nil, fmt.Errorf("cannot query for GetBuildsByRepositoryHash %s", err)
	}

	builds := make([]Build, 0)
	for rows.Next() {
		build := Build{}
		err := rows.Scan(&build.UID,
			&build.Status,
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
		build._exists = true
		builds = append(builds, build)
	}

	return &builds, nil
}
func (d *PostgresDBRepository) GetBuildBy(repositoryHash, commitHash string) (*Build, error) {
	row := d.connection.QueryRow("select * from public.build where repository_hash=$1 and commit_hash=$2", repositoryHash, commitHash)

	build := Build{}
	err := row.Scan(&build.UID,
		&build.RepositoryHash,
		&build.CommitHash,
		&build.PipelineJSON,
		&build.Logs,
		&build.Status,
		&build.ArtifactsUID,
		&build.CreatedAt,
		&build.FinishAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetBuildByCommitHash scan Build object failed %s", err)
	}
	build._exists = true

	return &build, nil

}
func (d *PostgresDBRepository) SetLastBuildCommit(key string, commitHash string) error {
	lastCommit := &KvBuildLastCommit{
		Key:             key,
		ValueHashCommit: commitHash,
		CreatedAt:       time.Now().Unix(),
	}
	err := lastCommit.Insert(d.connection)
	if err != sql.ErrNoRows {
		return err
	}

	return nil

}
func (d *PostgresDBRepository) GetLastBuildCommit(key string) (commitHash *string, createOn *int64, error error) {
	row := d.connection.QueryRow("select * from public.kv_build_last_commit where key=$1 order by id desc", key)

	lastCommit := KvBuildLastCommit{}
	err := row.Scan(&lastCommit.ID,
		&lastCommit.Key,
		&lastCommit.ValueHashCommit,
		&lastCommit.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("GetLastBuildCommit scan KvBuildLastCommit object failed %s", err)
	}
	lastCommit._exists = true

	return &lastCommit.ValueHashCommit, &lastCommit.CreatedAt, nil
}

func getRepositories(rows *sql.Rows) *[]Repository {
	repositories := make([]Repository, 0)
	for rows.Next() {
		repo := Repository{}
		rows.Scan(&repo.RepoHash, &repo.RepositoryURL, &repo.DefaultBranch, &repo.Name, &repo.CreatedAt, &repo.ProjectUID)
		repo._exists = true

		repositories = append(repositories, repo)
	}

	return &repositories
}

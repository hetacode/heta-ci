package handlers

import "github.com/hetacode/heta-ci/controller/app"

type FileCategory string

const (
	RepoFileCategory      FileCategory = "repo"
	ArtifactsFileCategory              = "artifacts"
)

type Handlers struct {
	Controller *app.Controller
}

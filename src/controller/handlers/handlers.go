package handlers

import "github.com/hetacode/heta-ci/controller/utils"

type FileCategory string

const (
	RepoFileCategory      FileCategory = "repo"
	ArtifactsFileCategory              = "artifacts"
)

type Handlers struct {
	Controller *utils.Controller
}

package app

import (
	"os"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/handlers"
	"github.com/hetacode/heta-ci/agent/services"
	"github.com/hetacode/heta-ci/agent/utils"
)

type App struct {
	Config               *utils.Config
	EventsHandlerManager *goeh.EventsHandlerManager
	MessagingService     *handlers.MessagingServiceHandler
	ScriptsHostDir       string
	ArtifactsHostDir     string
	ArtifactsHostInDir   string
	ArtifactsHostOutDir  string
	CodeHostDir          string
	ArtifactsService     *services.ArtifactsService
	RepositoryService    *services.RepositoryService
}

const controllerEndpoint = "http://localhost:5080"

func NewApp() *App {
	pwd, _ := os.Getwd()

	return &App{
		Config:               utils.NewConfig(),
		EventsHandlerManager: goeh.NewEventsHandlerManager(),
		ArtifactsService:     services.NewArtifactsService(controllerEndpoint),
		RepositoryService:    services.NewRepositoryService(controllerEndpoint),
		ScriptsHostDir:       pwd + "/scripts",
		CodeHostDir:          pwd + "/code",
		ArtifactsHostDir:     pwd + "/artifacts",
		ArtifactsHostInDir:   pwd + "/artifacts/in",
		ArtifactsHostOutDir:  pwd + "/artifacts/out",
	}
}

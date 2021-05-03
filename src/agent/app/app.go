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
	ArtifactsService     *services.ArtifactsService
}

func NewApp() *App {
	pwd, _ := os.Getwd()

	return &App{
		Config:               utils.NewConfig(),
		EventsHandlerManager: goeh.NewEventsHandlerManager(),
		ArtifactsService:     services.NewArtifactsService("http://localhost:5080"),
		ScriptsHostDir:       pwd + "/scripts",
		ArtifactsHostDir:     pwd + "/artifacts",
		ArtifactsHostInDir:   pwd + "/artifacts/in",
		ArtifactsHostOutDir:  pwd + "/artifacts/out",
	}
}

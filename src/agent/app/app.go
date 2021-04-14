package app

import (
	"os"

	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/handlers"
	"github.com/hetacode/heta-ci/agent/utils"
)

type App struct {
	Config               *utils.Config
	EventsHandlerManager *goeh.EventsHandlerManager
	MessagingService     *handlers.MessagingServiceHandler
	ScriptsHostDir       string
	ArtifactsHostDir     string
}

func NewApp() *App {
	pwd, _ := os.Getwd()

	return &App{
		Config:               utils.NewConfig(),
		EventsHandlerManager: goeh.NewEventsHandlerManager(),
		ScriptsHostDir:       pwd + "/scripts",
		ArtifactsHostDir:     pwd + "/artifacts",
	}
}

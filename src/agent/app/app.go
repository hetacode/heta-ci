package app

import (
	goeh "github.com/hetacode/go-eh"
	"github.com/hetacode/heta-ci/agent/utils"
)

type App struct {
	Config               *utils.Config
	EventsHandlerManager *goeh.EventsHandlerManager
}

func NewApp() *App {
	return &App{
		EventsHandlerManager: goeh.NewEventsHandlerManager(),
	}
}

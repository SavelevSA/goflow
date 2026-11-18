package main

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"

	"goflow/internal/ui"
)

func main() {
	app.Route("/", ui.NewDashboard)
	app.RunWhenOnBrowser()
	serve()
}

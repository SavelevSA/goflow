package ui

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type dashboard struct {
	app.Compo
	wasmReady bool
}

func NewDashboard() app.Composer {
	return &dashboard{}
}

func (d *dashboard) Render() app.UI {
	status := "Проверка WebAssembly"
	if d.wasmReady {
		status = "WebAssembly работает"
	}

	return app.Main().Body(
		app.H1().Text("GoFlow"),
		app.Button().Text(status).OnClick(func(_ app.Context, _ app.Event) {
			d.wasmReady = true
		}),
	)
}

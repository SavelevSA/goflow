package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"

	"goflow/internal/model"
)

type dashboard struct {
	snapshot model.Snapshot
	app.Compo
	wasmReady bool
	loadError string

	requestCtx context.Context
	cancel     context.CancelFunc

	commandError string
	commandBusy  bool
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
		app.P().Text(fmt.Sprintf("Перекрёстков: %d", len(d.snapshot.Intersections))),
		d.renderCity(),
		app.P().Text(d.loadError),
		app.Button().Text(status).OnClick(func(_ app.Context, _ app.Event) {
			d.wasmReady = true
		}),
	)
}

func (d *dashboard) renderCity() app.UI {
	var elements []app.UI

	for _, position := range []int{25, 75} { //четыре прямоугольника дороги
		elements = append(elements,
			app.Div().Style(
				"position", "absolute",
			).Style(
				"left", fmt.Sprintf("%d%%", position),
			).Style(
				"top", "0",
			).Style(
				"width", "12%",
			).Style(
				"height", "100%",
			).Style(
				"background", "#475569",
			).Style(
				"transform", "translateX(-50%)",
			),
		)

		elements = append(elements,
			app.Div().
				Style("position", "absolute").
				Style("left", "0").
				Style("top", fmt.Sprintf("%d%%", position)).
				Style("width", "100%").
				Style("height", "12%").
				Style("background", "#475569").
				Style("transform", "translateY(-50%)"),
		)

	}

	for _, intersection := range d.snapshot.Intersections { //подпись перекрёстков

		phaseText := "неизвестная фаза"
		if intersection.Phase == model.PhaseNS {
			phaseText = "зелёный: С–Ю"
		} else if intersection.Phase == model.PhaseEW {
			phaseText = "зелёный: В-З"
		}
		elements = append(elements,
			app.Div().
				Text(fmt.Sprintf("№%d %s", intersection.ID, phaseText)).
				Style("position", "absolute").
				Style("left", fmt.Sprintf("%g%%", intersection.X)).
				Style("top", fmt.Sprintf("%g%%", intersection.Y)).
				Style("transform", "translate(-50%, -50%)").
				Style("background", "white").
				Style("padding", "8px").
				Style("border-radius", "6px"),
		)
	}

	for _, vehicle := range d.snapshot.Vehicles {
		elements = append(elements,
			app.Div().
				Text(fmt.Sprintf("%d", vehicle.ID)).
				Style("position", "absolute").
				Style("left", fmt.Sprintf("%g%%", vehicle.X)).
				Style("top", fmt.Sprintf("%g%%", vehicle.Y)).
				Style("transform", "translate(-50%, -50%)").
				Style("width", "24px").
				Style("height", "16px").
				Style("background", "blue").
				Style("color", "white").
				Style("font-size", "12px").
				Style("text-align", "center"),
		)
	}

	return app.Div(). //собирает все в один контейнер карты
				Style("position", "relative").
				Style("width", "100%").
				Style("max-width", "600px").
				Style("aspect-ratio", "1").
				Style("background", "#bbf7d0").
				Body(elements...)
}

func (d *dashboard) loadSnapshot(ctx app.Context) {
	ctx.Async(func() {
		snapshot, err := fetchSnapshot(d.requestCtx)

		ctx.Dispatch(func(ctx app.Context) {
			ctx.After(300*time.Millisecond, d.loadSnapshot)
			if err != nil {
				d.loadError = err.Error()
				return
			}
			d.loadError = ""
			d.snapshot = snapshot
		})
	})

}

func (d *dashboard) OnDismount() {
	if d.cancel != nil {
		d.cancel()
	}
}

func (d *dashboard) OnMount(ctx app.Context) {
	d.requestCtx, d.cancel = context.WithCancel(context.Background())
	d.loadSnapshot(ctx)
}

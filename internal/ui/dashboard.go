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

	commandVersion int
}

func NewDashboard() app.Composer {
	return &dashboard{}
}

func (d *dashboard) Render() app.UI {
	status := "Проверка WebAssembly"
	if d.wasmReady {
		status = "WebAssembly работает"
	}

	pauseText := "Пауза"
	if d.snapshot.Paused {
		pauseText = "Продолжить"
	}

	return app.Main().Body(
		app.H1().Text("GoFlow"),
		app.P().Text(fmt.Sprintf("Перекрёстков: %d", len(d.snapshot.Intersections))),

		app.Button().
			Text(pauseText).
			Disabled(d.commandBusy).
			OnClick(func(ctx app.Context, _ app.Event) {
				payload := struct {
					Paused bool `json:"paused"`
				}{
					Paused: !d.snapshot.Paused,
				}

				d.runCommand(ctx, "/api/pause", payload)
			}),

		app.P().Text(fmt.Sprintf("Режим: %s", d.snapshot.Mode)),

		app.Button().
			Text("Fixed").
			Disabled(d.commandBusy).
			OnClick(func(ctx app.Context, _ app.Event) {
				payload := struct {
					Mode model.Mode `json:"mode"`
				}{
					Mode: model.ModeFixed,
				}

				d.runCommand(ctx, "/api/mode", payload)
			}),

		app.Button().
			Text("Adaptive").
			Disabled(d.commandBusy).
			OnClick(func(ctx app.Context, _ app.Event) {
				payload := struct {
					Mode model.Mode `json:"mode"`
				}{
					Mode: model.ModeAdaptive,
				}

				d.runCommand(ctx, "/api/mode", payload)
			}),

		app.P().Text(fmt.Sprintf("Плотность: %d", d.snapshot.Density)),

		app.Button().
			Text("−").
			Disabled(d.commandBusy || d.snapshot.Density <= 1).
			OnClick(func(ctx app.Context, _ app.Event) {
				d.changeDensity(ctx, d.snapshot.Density-1)
			}),

		app.Button().
			Text("+").
			Disabled(d.commandBusy || d.snapshot.Density >= 5).
			OnClick(func(ctx app.Context, _ app.Event) {
				d.changeDensity(ctx, d.snapshot.Density+1)
			}),

		app.Button().
			Text("Сбросить").
			Disabled(d.commandBusy).
			OnClick(func(ctx app.Context, _ app.Event) {
				payload := struct {
					Seed int64 `json:"seed"`
				}{
					Seed: d.snapshot.Seed,
				}

				d.runCommand(ctx, "/api/reset", payload)
			}),

		d.renderCity(),

		d.renderStatistics(),

		app.P().Text(d.loadError),
		app.P().Text(d.commandError),
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

	segments := []struct {
		start  int
		length int
	}{
		{start: 0, length: 19},
		{start: 31, length: 38},
		{start: 81, length: 19},
	}

	for _, position := range []int{25, 75} {
		for _, segment := range segments {
			elements = append(elements,
				app.Div().
					Style("position", "absolute").
					Style("left", fmt.Sprintf("%d%%", position)).
					Style("top", fmt.Sprintf("%d%%", segment.start)).
					Style("height", fmt.Sprintf("%d%%", segment.length)).
					Style("border-left", "2px dashed white").
					Style("transform", "translateX(-50%)"),

				app.Div().
					Style("position", "absolute").
					Style("left", fmt.Sprintf("%d%%", segment.start)).
					Style("top", fmt.Sprintf("%d%%", position)).
					Style("width", fmt.Sprintf("%d%%", segment.length)).
					Style("border-top", "2px dashed white").
					Style("transform", "translateY(-50%)"),
			)
		}
	}

	for _, intersection := range d.snapshot.Intersections {
		x := intersection.X
		y := intersection.Y
		nsColor := "red"
		ewColor := "red"

		if intersection.Phase == model.PhaseNS {
			nsColor = "lime"
		} else if intersection.Phase == model.PhaseEW {
			ewColor = "lime"
		}

		elements = append(elements,
			stopLine(x-6, y, true),
			stopLine(x+6, y-6, true),
			stopLine(x-6, y-6, false),
			stopLine(x, y+6, false),
			trafficLight(x-6, y, ewColor),
			trafficLight(x+6, y, ewColor),
			trafficLight(x, y-6, nsColor),
			trafficLight(x, y+6, nsColor),
		)
	}

	for _, vehicle := range d.snapshot.Vehicles {

		angle := 0

		switch vehicle.Direction {
		case model.North:
			angle = -90
		case model.South:
			angle = 90
		case model.West:
			angle = 180
		case model.East:
			angle = 0
		}

		elements = append(elements,
			app.Div().
				Style("position", "absolute").
				Style("left", fmt.Sprintf("%g%%", vehicle.X)).
				Style("top", fmt.Sprintf("%g%%", vehicle.Y)).
				Style("transform", fmt.Sprintf(
					"translate(-50%%, -50%%) rotate(%ddeg)", angle,
				)).
				Style("width", "6%").
				Style("height", "3%").
				Style("background", "blue").
				Style("color", "white").
				Style("font-size", "12px").
				Style("text-align", "center").
				Body(
					app.Span().Text(fmt.Sprintf("%d", vehicle.ID)),

					app.Span().
						Style("position", "absolute").
						Style("right", "0").
						Style("top", "10%").
						Style("width", "15%").
						Style("height", "25%").
						Style("background", "yellow"),

					app.Span().
						Style("position", "absolute").
						Style("right", "0").
						Style("bottom", "10%").
						Style("width", "15%").
						Style("height", "25%").
						Style("background", "yellow"),
				),
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
	version := d.commandVersion
	ctx.Async(func() {
		snapshot, err := fetchSnapshot(d.requestCtx)

		ctx.Dispatch(func(ctx app.Context) {
			ctx.After(300*time.Millisecond, d.loadSnapshot)
			if d.commandBusy || version != d.commandVersion {
				return
			}
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

func (d *dashboard) runCommand(
	ctx app.Context,
	path string,
	payload any,
) {
	if d.commandBusy {
		return
	}
	d.commandVersion++
	d.commandBusy = true
	d.commandError = ""

	ctx.Async(func() {
		snapshot, err := sendCommand(d.requestCtx, path, payload)

		ctx.Dispatch(func(ctx app.Context) {
			d.commandVersion++
			d.commandBusy = false

			if err != nil {
				d.commandError = err.Error()
				return
			}

			d.snapshot = snapshot
		})
	})
}

func (d *dashboard) changeDensity(ctx app.Context, density int) {
	payload := struct {
		Density int `json:"density"`
	}{
		Density: density,
	}

	d.runCommand(ctx, "/api/density", payload)
}

func (d *dashboard) renderStatistics() app.UI {
	stats := d.snapshot.Statistics

	return app.Div().Body(
		app.H2().Text("Статистика"),
		app.P().Text(fmt.Sprintf(
			"Активных машин: %d", stats.ActiveVehicles,
		)),
		app.P().Text(fmt.Sprintf(
			"Завершённых поездок: %d", stats.TotalCompleted,
		)),
		app.P().Text(fmt.Sprintf(
			"В очереди: %d", stats.TotalQueue,
		)),
		app.P().Text(fmt.Sprintf(
			"Среднее ожидание: %.1f с", stats.AverageWaitingTime,
		)),

		app.H2().Text("Сравнение режимов"),

		app.Div().
			Style("display", "grid").
			Style("grid-template-columns", "repeat(auto-fit, minmax(220px, 1fr))").
			Style("gap", "16px").
			Body(
				renderModePanel("Fixed", d.snapshot.ModeStats.Fixed),
				renderModePanel("Adaptive", d.snapshot.ModeStats.Adaptive),
			),
	)
}

func renderModePanel(name string, stats model.ModeStatistics) app.UI {
	return app.Div().
		Style("padding", "16px").
		Body(
			app.H3().Text(name),

			app.P().Text(fmt.Sprintf(
				"Время работы: %.1f с",
				float64(stats.ActiveTicks)*0.25,
			)),

			app.P().Text(fmt.Sprintf(
				"Завершённых пересечений: %d",
				stats.CompletedCrossings,
			)),

			app.P().Text(fmt.Sprintf(
				"Ожидавших автомобилей: %d",
				stats.WaitingVehicles,
			)),

			app.P().Text(fmt.Sprintf(
				"Суммарное ожидание: %.1f с",
				stats.WaitingSeconds,
			)),

			app.P().Text(fmt.Sprintf(
				"Среднее ожидание: %.1f с",
				stats.AverageWaitingTime,
			)),
		)
}

func stopLine(x, y float64, vertical bool) app.UI {
	line := app.Div().
		Style("position", "absolute").
		Style("left", fmt.Sprintf("%g%%", x)).
		Style("top", fmt.Sprintf("%g%%", y)).
		Style("background", "white")

	if vertical {
		return line.
			Style("width", "2px").
			Style("height", "6%").
			Style("transform", "translateX(-50%)")
	}

	return line.
		Style("width", "6%").
		Style("height", "2px").
		Style("transform", "translateY(-50%)")
}

func trafficLight(x, y float64, color string) app.UI {
	return app.Div().
		Style("position", "absolute").
		Style("left", fmt.Sprintf("%g%%", x)).
		Style("top", fmt.Sprintf("%g%%", y)).
		Style("width", "12px").
		Style("height", "12px").
		Style("border-radius", "50%").
		Style("background", color).
		Style("border", "1px solid black").
		Style("transform", "translate(-50%, -50%)")
}

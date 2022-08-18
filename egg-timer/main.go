package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type C = layout.Context
type D = layout.Dimensions

var progress float32
var progressIncrement chan float32

func draw(w *app.Window) error {
	//ops are the operations from UI
	var ops op.Ops
	var startButton widget.Clickable // startbutton is  clickable widget
	var boiling bool
	var boilDuration = float32(100)
	var boilDurationInput widget.Editor

	th := material.NewTheme(gofont.Collection()) // th defines the material design style
	// listen for events
	for {
		select {
		case p := <-progressIncrement:
			if boiling && progress < 1 {
				boilRemain := (1 - progress) * boilDuration
				inputStr := fmt.Sprintf("%.2f", math.Round(float64(boilRemain)*10)/10)
				boilDurationInput.SetText(inputStr)
				progress += p
				w.Invalidate()
			}
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
				// sent when the app should re-render
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				if startButton.Clicked() {
					boiling = !boiling
				}

				layout.Flex{
					Axis:    layout.Vertical,
					Spacing: layout.SpaceStart,
				}.Layout(gtx,
					// The Egg
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							var eggPath clip.Path
							op.Offset(image.Pt(200, 150)).Add(gtx.Ops)
							eggPath.Begin(gtx.Ops)

							// https://observablehq.com/@toja/egg-curve
							for deg := 0.0; deg <= 360; deg++ {
								rad := deg / 360 * 2 * math.Pi
								cosT := math.Cos(rad)
								sinT := math.Sin(rad)

								a := 110.0
								b := 150.0
								d := 20.0

								x := a * sinT
								y := -(math.Sqrt(b*b-d*d*sinT*sinT) + d*cosT) * cosT
								p := f32.Pt(float32(x), float32(y))
								eggPath.LineTo(p)
							}

							eggPath.Close()
							eggArea := clip.Outline{Path: eggPath.End()}.Op()
							clr := color.NRGBA{R: 255, G: uint8(239 * (1 - progress)), B: uint8(174 * (1 - progress)), A: 255}
							paint.FillShape(gtx.Ops, clr, eggArea)

							d := image.Point{Y: 375}
							return layout.Dimensions{Size: d}
						},
					),
					// The input box
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {

							boilDurationInput.SingleLine = true
							boilDurationInput.Alignment = text.Middle
							boilDurationInput.MaxLen = 4
							ed := material.Editor(th, &boilDurationInput, "Time in Secs")

							margins := layout.Inset{
								Top:    unit.Dp(0),
								Right:  unit.Dp(10),
								Bottom: unit.Dp(40),
								Left:   unit.Dp(0),
							}
							border := widget.Border{
								Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								CornerRadius: unit.Dp(3),
								Width:        unit.Dp(1),
							}
							return margins.Layout(gtx,
								func(gt layout.Context) layout.Dimensions {
									return border.Layout(gtx, ed.Layout)
								},
							)
						},
					),
					// The progress bar
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							bar := material.ProgressBar(th, progress)
							return bar.Layout(gtx)
						},
					),
					// The button
					layout.Rigid(
						func(gtx layout.Context) layout.Dimensions {
							margins := layout.Inset{
								Top:    unit.Dp(25),
								Bottom: unit.Dp(25),
								Right:  unit.Dp(35),
								Left:   unit.Dp(35),
							}
							return margins.Layout(gtx,
								func(gtx layout.Context) layout.Dimensions {
									var text = "Start"
									if boiling {
										text = "stop"
									} else if boiling && progress >= 1 {
										text = "Finished"
									}
									btn := material.Button(th, &startButton, text)
									return btn.Layout(gtx)
								},
							)
						},
					),
					layout.Rigid(
						layout.Spacer{Height: unit.Dp(25)}.Layout,
					),
				)

				e.Frame(gtx.Ops)
			}
		}
	}
}

func main() {
	progressIncrement = make(chan float32)
	go func() {
		for {
			time.Sleep(time.Second / 25)
			progressIncrement <- 0.004
		}
	}()

	go func() {
		//create a new window
		w := app.NewWindow(
			app.Title("Egg Timer"),
			app.Size(unit.Dp(400), unit.Dp(600)),
		)
		if err := draw(w); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()
	app.Main()
}

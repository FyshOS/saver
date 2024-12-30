package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/FyshOS/saver"
)

func main() {
	a := app.NewWithID("com.fyshos.fyshsaver")

	s := saver.NewScreenSaver(a.Quit)
	s.ClockFormat = a.Preferences().StringWithFallback("clockformatting", "12h")
	s.Label = a.Preferences().StringWithFallback("fysh.label", "FyshOS")
	s.Lock = true

	s.ShowWindow()
	a.Run()
}

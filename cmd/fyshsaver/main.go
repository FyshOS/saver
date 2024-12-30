package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/FyshOS/saver"
)

func main() {
	a := app.NewWithID("com.fyshos.fyshsaver")

	s := &saver.ScreenSaver{
		ClockFormat: a.Preferences().StringWithFallback("clockformatting", "12h"),
		Label:       a.Preferences().StringWithFallback("fysh.label", "FyshOS"),
		Lock:        true,
	}

	s.ShowWindow(a)
	a.Run()
}

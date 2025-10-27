package main

import (
	"flag"

	"fyne.io/fyne/v2/app"

	"github.com/FyshOS/saver"
)

func main() {
	a := app.NewWithID("com.fyshos.fyshsaver")

	lock := flag.Bool("lock", false, "Lock the screen")
	label := flag.String("label", "(clock)", "Label to display")
	flag.Parse()

	s := saver.NewScreenSaver(a.Quit)
	s.ClockFormat = a.Preferences().StringWithFallback("clockformatting", "12h")
	s.Label = *label
	s.Lock = *lock
	s.LockImmediately = true // when using CLI always be immediate

	s.ShowWindows()
	a.Run()
}

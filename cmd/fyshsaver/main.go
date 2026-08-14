package main

import (
	"flag"

	"fyne.io/fyne/v2/app"

	"github.com/FyshOS/saver"
)

func main() {
	a := app.NewWithID("com.fyshos.fyshsaver")
	a.Settings().SetTheme(newDarkBackgroundTheme())

	lock := flag.Bool("lock", false, "Lock the screen")
	delay := flag.Bool("lock-delay", false, "Delay the screen lock (by 3 sec)")
	label := flag.String("label", "(clock)", "Label to display")
	suspending := flag.Bool("suspending", false, "The machine is going to sleep, so start with animations stopped")
	flag.Parse()

	s := saver.NewScreenSaver(a.Quit)
	s.ClockFormat = a.Preferences().StringWithFallback("clockformatting", "12h")
	s.Label = *label
	s.Lock = *lock
	s.LockImmediately = !*delay // when using CLI default to immediate lock
	s.Suspending = *suspending

	s.ShowWindows()
	a.Run()
}

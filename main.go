package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func main() {
	a := app.NewWithID("com.fyshos.screensaver")
	w := a.NewWindow("Screensaver")
	w.Resize(fyne.NewSize(500, 350))

	ico1 := canvas.NewImageFromFile("fish.png")
	ico1.Resize(fyne.NewSquareSize(96))
	l1 := &moveLayout{xInc: 2, yInc: 2}

	ico2 := canvas.NewImageFromFile("fish.png")
	ico2.Resize(fyne.NewSquareSize(96))
	ico2.Move(fyne.NewPos(400, 100))
	l2 := &moveLayout{xInc: 2.2, yInc: 2.2, invertX: true}

	ico3 := canvas.NewImageFromFile("fish.png")
	ico3.Resize(fyne.NewSquareSize(96))
	l3 := &moveLayout{xInc: 1.5, yInc: 1.5}
	ico3.Move(fyne.NewPos(220, 310))
	l3.invertY = true

	ico4 := canvas.NewImageFromFile("fish.png")
	ico4.Resize(fyne.NewSquareSize(128))
	l4 := &moveLayout{xInc: 1, yInc: 1, invertX: true}
	ico4.Move(fyne.NewPos(450, 200))

	ico5 := canvas.NewImageFromFile("fish.png")
	ico5.Resize(fyne.NewSquareSize(128))
	l5 := &moveLayout{xInc: 1, yInc: 1, invertY: true}
	ico5.Move(fyne.NewPos(150, 300))

	label := a.Preferences().StringWithFallback("fysh.label", "FyshOS")

	txt := canvas.NewText(label, theme.Color(theme.ColorNameForeground))
	//if label == "(clock)" {
	format := fyne.CurrentApp().Preferences().StringWithFallback("clockformatting", "12h")
	go clockText(txt, format)
	//}

	txt.TextSize = 84
	txt.Resize(txt.MinSize())
	l6 := &moveLayout{xInc: 1, yInc: 1}
	w.SetContent(container.NewStack(container.New(l6, txt), container.New(l5, ico5), container.New(l4, ico4), container.New(l3, ico3), container.New(l2, ico2), container.New(l1, ico1)))

	go l1.run()
	go l2.run()
	go l3.run()
	go l4.run()
	go l5.run()
	go l6.run()

	w.Canvas().SetOnTypedRune(func(r rune) {
		startedInput(a)
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		startedInput(a)
	})

	w.SetPadded(false)
	w.SetFullScreen(true)
	w.ShowAndRun()
}

func formattedTime(format string) string { // matching the desktop format
	if format == "12h" {
		return time.Now().Format("03:04pm")
	}

	return time.Now().Format("15:04")
}

func clockText(t *canvas.Text, format string) {
	for {
		txt := formattedTime(format)
		t.Text = txt
		t.Resize(t.MinSize())

		time.Sleep(time.Second * 10) // don't refresh too fast but don't lag more than 10 sec
	}
}

func startedInput(a fyne.App) {
	a.Quit() // TODO password support
}

type moveLayout struct {
	size fyne.Size
	objs []fyne.CanvasObject

	invertX, invertY bool
	xInc, yInc       float32
}

func (m *moveLayout) Layout(o []fyne.CanvasObject, s fyne.Size) {
	m.size = s
	m.objs = o
}

func (m *moveLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSquareSize(300)
}

func (m *moveLayout) move() {
	o := m.objs[0]

	x, y := o.Position().Components()

	if m.invertX {
		x -= m.xInc
		if x < 0 {
			m.invertX = false
			x = 1
		}
	} else {
		x += m.xInc
		if x >= m.size.Width-o.Size().Width {
			m.invertX = true
			x = m.size.Width - o.Size().Width - m.xInc*2
		}
	}

	if m.invertY {
		y -= m.yInc
		if y < 0 {
			m.invertY = false
			y = 1
		}
	} else {
		y += m.yInc
		if y >= m.size.Height-o.Size().Height {
			m.invertY = true
			y = m.size.Height - o.Size().Height - m.yInc*2
		}
	}

	o.Move(fyne.NewPos(x, y))
}

func (m *moveLayout) run() {
	for {
		time.Sleep(time.Second / 60) // TODO use animation

		if len(m.objs) == 0 || m.size.Width == 0 {
			continue
		}
		m.move()
	}
}

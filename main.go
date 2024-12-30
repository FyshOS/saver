package main

import (
	"embed"
	"fmt"
	"image"
	"image/png"
	"time"

	"github.com/nfnt/resize"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

const (
	clockLabelKey = "(clock)"
	frameCount    = 5
)

var (
	//go:embed "frames"
	frames embed.FS

	fyshes [frameCount]image.Image
)

type ScreenSaver struct {
	Label       string
	Lock        bool
	ClockFormat string
}

func (s *ScreenSaver) showClock() bool {
	return s.Label == clockLabelKey
}

func main() {
	a := app.NewWithID("com.fyshos.screensaver")
	w := a.NewWindow("Screensaver")
	w.Resize(fyne.NewSize(500, 350))

	s := &ScreenSaver{
		ClockFormat: a.Preferences().StringWithFallback("clockformatting", "12h"),
		Label:       a.Preferences().StringWithFallback("fysh.label", "FyshOS"),
		Lock:        true,
	}

	w.SetContent(s.MakeUI(a, w))
	w.Canvas().SetOnTypedRune(func(r rune) {
		startedInput(a)
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		startedInput(a)
	})

	w.SetPadded(false)
	w.SetFullScreen(true)
	a.Lifecycle().SetOnEnteredForeground(func() {
		hideCursor(w)
	})

	w.ShowAndRun()
}

func (s *ScreenSaver) MakeUI(a fyne.App, w fyne.Window) fyne.CanvasObject {
	for i := 0; i < frameCount; i++ {
		name := fmt.Sprintf("fysh%d.png", i)
		frame, _ := frames.Open("frames/" + name)

		full, _ := png.Decode(frame)
		fyshes[i] = resize.Resize(uint(128), uint(128), full, resize.Lanczos3)
		_ = frame.Close()
	}

	ico1 := newFysh(96)
	l1 := &moveLayout{xInc: 2, yInc: 2}

	ico2 := newFysh(96)
	ico2.Move(fyne.NewPos(400, 100))
	l2 := &moveLayout{xInc: 2.2, yInc: 2.2, invertX: true}

	ico3 := newFysh(96)
	l3 := &moveLayout{xInc: 1.5, yInc: 1.5}
	ico3.Move(fyne.NewPos(220, 310))
	l3.invertY = true

	ico4 := newFysh(128)
	l4 := &moveLayout{xInc: 1, yInc: 1, invertX: true}
	ico4.Move(fyne.NewPos(450, 200))

	ico5 := newFysh(128)
	l5 := &moveLayout{xInc: 1, yInc: 1, invertY: true}
	ico5.Move(fyne.NewPos(150, 300))

	txt := canvas.NewText(s.Label, theme.Color(theme.ColorNameForeground))
	if s.showClock() {
		go clockText(txt, s.ClockFormat)
	}

	txt.TextSize = 84
	txt.Resize(txt.MinSize())
	l6 := &moveLayout{xInc: 1, yInc: 1}

	go l1.run()
	go l2.run()
	go l3.run()
	go l4.run()
	go l5.run()
	go l6.run()

	return container.NewStack(
		&cursorCapture{moved: func() {
			startedInput(a)
		}},
		container.New(l6, txt),
		container.New(l5, ico5),
		container.New(l4, ico4),
		container.New(l3, ico3),
		container.New(l2, ico2),
		container.New(l1, ico1))
}

var fyshCount = 0

func newFysh(size int) *canvas.Image {
	ico := &canvas.Image{}
	ico.Resize(fyne.NewSquareSize(float32(size)))
	id := fyshCount % 5
	fyshCount++

	go func() {
		for {
			ico.Image = fyshes[id]
			ico.Refresh()

			id++
			if id >= 5 {
				id = 0
			}
			time.Sleep(time.Millisecond * 800)
		}
	}()

	return ico
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

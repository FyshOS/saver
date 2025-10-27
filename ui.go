package saver

import (
	"embed"
	"fmt"
	"image"
	"image/png"
	"log"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/nfnt/resize"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
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
	Label                 string
	Lock, LockImmediately bool
	ClockFormat           string

	OnUnlocked func()
	started    time.Time
	wins       []fyne.Window
}

func NewScreenSaver(onUnlocked func()) *ScreenSaver {
	s := &ScreenSaver{OnUnlocked: onUnlocked}
	initCursor()
	s.started = time.Now()
	return s
}

func (s *ScreenSaver) showClock() bool {
	return s.Label == clockLabelKey
}

func (s *ScreenSaver) makeWindow(content func(fyne.Window) fyne.CanvasObject) fyne.Window {
	w := fyne.CurrentApp().Driver().(desktop.Driver).CreateSplashWindow()
	w.SetPadded(false)
	w.Resize(fyne.NewSize(500, 350))

	w.SetContent(content(w))
	w.Canvas().SetOnTypedRune(func(r rune) {
		s.startedInput(w)
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		s.startedInput(w)
	})

	return w
}

func (s *ScreenSaver) ShowWindow() {
	w := s.makeWindow(s.MakeUI)
	go func() {
		time.Sleep(time.Millisecond * 250)
		hideCursor(w)
	}()

	s.wins = []fyne.Window{w}
	w.SetFullScreen(true)
	w.Show()
}

func (s *ScreenSaver) ShowWindows() {
	w := s.makeWindow(s.MakeUI)
	go func() {
		time.Sleep(time.Millisecond * 250)
		hideCursor(w)
	}()

	s.wins = []fyne.Window{w}
	w.Show()

	go func() {
		conn, err := xgb.NewConn()
		if err != nil {
			log.Println("Failed to get X11 connection", err)
			return
		}

		screens := glfw.GetMonitors()
		for i, scr := range screens {
			if i > 0 {
				// a horrible hack until we track down the concurrent window configure deadlock
				time.Sleep(time.Millisecond * 500)

				w2 := fyne.CurrentApp().Driver().(desktop.Driver).CreateSplashWindow()
				w2.SetContent(s.MakeUI(w2))

				w2.Show()
				w = w2
				s.wins = append(s.wins, w2)
			}

			w.(driver.NativeWindow).RunNative(func(ctx any) {
				xWin := ctx.(driver.X11WindowContext).WindowHandle
				mode := scr.GetVideoMode()

				x, y := scr.GetPos()
				go func() {
					for i := 0; i < 10; i++ {

						fyne.Do(func() {
							_ = xproto.ConfigureWindow(conn, xproto.Window(xWin), xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
								[]uint32{uint32(x - 1), uint32(y - 1), uint32(mode.Width + 2), uint32(mode.Height + 2)})
						})
						time.Sleep(time.Millisecond * 50)
					}
				}()
			})

		}
	}()
}

func (s *ScreenSaver) MakeUI(w fyne.Window) fyne.CanvasObject {
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
		txt.Text = formattedTime(s.ClockFormat)
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
			s.startedInput(w)
		}},
		container.New(l6, txt),
		container.New(l5, ico5),
		container.New(l4, ico4),
		container.New(l3, ico3),
		container.New(l2, ico2),
		container.New(l1, ico1))
}

func (s *ScreenSaver) unlock() {
	for _, w := range s.wins {
		if w != nil {
			w.Close()
		}
	}
	if fn := s.OnUnlocked; fn != nil {
		fn()
		return
	}

	fyne.CurrentApp().Quit()
}

var fyshCount = 0

func newFysh(size int) *canvas.Image {
	ico := &canvas.Image{}
	ico.Resize(fyne.NewSquareSize(float32(size)))
	id := fyshCount % 5
	fyshCount++

	go func() {
		for {
			fyne.Do(func() {
				ico.Image = fyshes[id]
				ico.Refresh()
			})

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
		return time.Now().Format("3:04pm")
	}

	return time.Now().Format("15:04")
}

func clockText(t *canvas.Text, format string) {
	oldTime := ""
	for {
		time.Sleep(time.Second)

		txt := formattedTime(format)
		if txt != oldTime {
			oldTime = txt
			fyne.DoAndWait(func() {
				t.Text = txt
				t.Resize(t.MinSize())
			})
		}
	}
}

func (s *ScreenSaver) startedInput(w fyne.Window) {
	if s.started.After(time.Now().Add(time.Millisecond * -200)) {
		return // something flickering as we start
	}
	if !s.Lock || (!s.LockImmediately && s.started.After(time.Now().Add(time.Second*-3))) {
		showCursor(w)
		s.unlock()
		return
	}

	showLogin(s.unlock, w)
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

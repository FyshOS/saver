package saver

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/nfnt/resize"
)

const (
	WindowTitle = "fyshsaver"

	clockLabelKey = "(clock)"
	frameCount    = 5

	fyshFrameTime   = time.Millisecond * 800
	animationPeriod = time.Minute
)

var (
	//go:embed "frames"
	frames embed.FS

	fyshes     [frameCount]image.Image
	fyshesOnce sync.Once
)

// loadFysh decodes the frames of the swimming fysh, once however many screens
// we are asked to fill.
func loadFysh() {
	fyshesOnce.Do(func() {
		for i := 0; i < frameCount; i++ {
			name := fmt.Sprintf("fysh%d.png", i)
			frame, _ := frames.Open("frames/" + name)

			full, _ := png.Decode(frame)
			fyshes[i] = resize.Resize(uint(128), uint(128), full, resize.Lanczos3)
			_ = frame.Close()
		}
	})
}

type ScreenSaver struct {
	Label                 string
	Lock, LockImmediately bool
	ClockFormat           string

	// Suspending marks that the machine is on its way to sleep as we start, so
	// the UI begins stopped and stays that way until the system wakes.
	Suspending bool

	ui CustomUI

	OnUnlocked func()
	started    time.Time

	winMu  sync.Mutex
	wins   []fyne.Window
	loaded bool

	asleep      atomic.Bool
	suspendStop func()
}

// prepare records the state we are starting in and begins listening for system
// suspend/resume, so the UI can stop its animations while asleep.
func (s *ScreenSaver) prepare() {
	initCursor()
	s.asleep.Store(s.Suspending)

	if s.suspendStop != nil {
		return
	}

	s.suspendStop = watchSuspend(func(sleeping bool) {
		s.asleep.Store(sleeping)
		if s.ui == nil {
			return
		}
		if sleeping {
			s.ui.Stop()
		} else {
			s.ui.Start()
		}
	})
}

// startUI begins the animations, unless we are showing because the machine is
// on its way to sleep - then they wait for the resume.
func (s *ScreenSaver) startUI() {
	if s.ui == nil || s.asleep.Load() {
		return
	}

	s.ui.Start()
}

func NewScreenSaver(onUnlocked func()) *ScreenSaver {
	s := &ScreenSaver{OnUnlocked: onUnlocked}
	s.started = time.Now()
	return s
}

func (s *ScreenSaver) showClock() bool {
	return s.Label == clockLabelKey
}

func (s *ScreenSaver) makeWindow() fyne.Window {
	var w fyne.Window
	if desk, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		w = desk.CreateSplashWindow()
	} else { // web or mobile
		w = fyne.CurrentApp().NewWindow("")
	}
	w.SetTitle(WindowTitle)

	w.SetPadded(false)
	w.Resize(fyne.NewSize(500, 350))
	w.SetContent(container.NewWithoutLayout())

	w.Canvas().SetOnTypedRune(func(r rune) {
		s.startedInput(w)
	})
	w.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		s.startedInput(w)
	})

	return w
}

func (s *ScreenSaver) registerWindow(w fyne.Window) {
	s.winMu.Lock()
	s.wins = append(s.wins, w)
	loaded := s.loaded
	s.winMu.Unlock()

	if loaded {
		w.SetContent(s.MakeUI(w))
		go fyne.Do(s.startUI) // don't delay show
	}
}

func (s *ScreenSaver) loadContent() {
	s.winMu.Lock()
	if s.loaded {
		s.winMu.Unlock()
		return
	}
	s.loaded = true
	wins := append([]fyne.Window(nil), s.wins...)
	s.winMu.Unlock()

	for _, w := range wins {
		w.SetContent(s.MakeUI(w))
		go fyne.Do(s.startUI) // don't delay show
	}
}

func (s *ScreenSaver) ShowWindow() {
	s.prepare()

	w := s.makeWindow()
	go func() {
		time.Sleep(time.Millisecond * 250)
		hideCursor(w)
	}()

	fyne.CurrentApp().Lifecycle().SetOnEnteredForeground(s.loadContent)
	s.registerWindow(w)
	w.SetFullScreen(true)
	w.Show()
}

func (s *ScreenSaver) ShowWindows() {
	s.prepare()

	w := s.makeWindow()
	go func() {
		time.Sleep(time.Millisecond * 250)
		hideCursor(w)
	}()

	fyne.CurrentApp().Lifecycle().SetOnEnteredForeground(s.loadContent)
	s.registerWindow(w)
	w.Show()

	go func() {
		conn, err := xgb.NewConn()
		if err != nil {
			log.Println("Failed to get X11 connection", err)
			return
		}

		screens := getMonitors()
		for i, scr := range screens {
			if i > 0 {
				// a horrible hack until we track down the concurrent window configure deadlock
				time.Sleep(time.Millisecond * 500)

				w2 := s.makeWindow()
				w2.Show()
				w = w2
				s.registerWindow(w2)
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

// saverScene is the animated content of one screen saver window. The saver asks
// for one per screen and they all start and stop together.
type saverScene struct {
	anims []*fyne.Animation

	clockLabel  *canvas.Text
	clockFormat string

	running bool
	clock   chan struct{}
}

// start runs this scene's animations, and its clock if it has one. It is a
// no-op if the scene is already running.
func (s *saverScene) start() {
	if s.running {
		return
	}
	s.running = true

	for _, a := range s.anims {
		a.Start()
	}
	if s.clockLabel != nil {
		s.clock = make(chan struct{})
		go clockText(s.clockLabel, s.clockFormat, s.clock)
	}
}

// stop halts everything start began, and may be started again afterwards.
func (s *saverScene) stop() {
	if !s.running {
		return
	}
	s.running = false

	for _, a := range s.anims {
		a.Stop()
	}
	if s.clock != nil {
		close(s.clock)
		s.clock = nil
	}
}

// defaultSaverUI tracks every scene it has built, as the screen saver asks for
// one per screen and they all start and stop together.
type defaultSaverUI struct {
	mu     sync.Mutex
	scenes []*saverScene
}

func (d *defaultSaverUI) MakeUI(s *ScreenSaver) fyne.CanvasObject {
	loadFysh()

	ico1, a1 := newFysh(96)
	l1 := &moveLayout{xInc: 2, yInc: 2}

	ico2, a2 := newFysh(96)
	ico2.Move(fyne.NewPos(400, 100))
	l2 := &moveLayout{xInc: 1.75, yInc: 1.75, invertX: true}

	ico3, a3 := newFysh(96)
	l3 := &moveLayout{xInc: 2.25, yInc: 2.25}
	ico3.Move(fyne.NewPos(220, 310))
	l3.invertY = true

	ico4, a4 := newFysh(128)
	l4 := &moveLayout{xInc: 1, yInc: 1, invertX: true}
	ico4.Move(fyne.NewPos(450, 200))

	ico5, a5 := newFysh(128)
	l5 := &moveLayout{xInc: 1.25, yInc: 1.25, invertY: true}
	ico5.Move(fyne.NewPos(150, 300))

	scene := &saverScene{}
	txt := canvas.NewText(s.Label, theme.Color(theme.ColorNameForeground))
	if s.showClock() {
		txt.Text = formattedTime(s.ClockFormat)
		scene.clockLabel, scene.clockFormat = txt, s.ClockFormat
	}

	txt.TextSize = 84
	txt.Resize(txt.MinSize())
	l6 := &moveLayout{xInc: 1.5, yInc: 1}

	scene.anims = []*fyne.Animation{a1, a2, a3, a4, a5,
		l1.animate(), l2.animate(), l3.animate(), l4.animate(), l5.animate(), l6.animate()}

	d.mu.Lock()
	d.scenes = append(d.scenes, scene)
	d.mu.Unlock()

	return container.NewStack(
		canvas.NewRectangle(color.Black),
		container.New(l6, txt),
		container.New(l5, ico5),
		container.New(l4, ico4),
		container.New(l3, ico3),
		container.New(l2, ico2),
		container.New(l1, ico1),
	)
}

func (s *ScreenSaver) MakeUI(w fyne.Window) fyne.CanvasObject {
	if s.ui == nil {
		s.ui = &defaultSaverUI{}
	}
	w.Canvas().SetOnTypedRune(func(_ rune) {
		s.startedInput(w)
	})
	w.Canvas().SetOnTypedKey(func(_ *fyne.KeyEvent) {
		s.startedInput(w)
	})
	ui := s.ui.MakeUI(s)
	w.SetOnClosed(s.ui.DestroyUI)

	return container.NewStack(
		&cursorCapture{moved: func() {
			s.startedInput(w)
		}},
		ui,
	)
}

// DestroyUI is called as the saver windows close, which happens to them all at
// once, so we tear down every scene the first time through.
func (d *defaultSaverUI) DestroyUI() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, scene := range d.scenes {
		scene.stop()
	}
	d.scenes = nil
}

// Stop implements CustomUI: it halts the animations while the system is asleep,
// so we are not drawing for a screen that is not there.
func (d *defaultSaverUI) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, scene := range d.scenes {
		scene.stop()
	}
}

// Start implements CustomUI, restarting the animations after wake.
func (d *defaultSaverUI) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, scene := range d.scenes {
		scene.start()
	}
}

func (s *ScreenSaver) unlock() {
	if s.suspendStop != nil {
		s.suspendStop()
		s.suspendStop = nil
	}

	s.winMu.Lock()
	wins := append([]fyne.Window(nil), s.wins...)
	s.winMu.Unlock()
	for _, w := range wins {
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

// newFysh returns a fysh of the given size alongside the animation that swims
// it, which the caller starts. Each fysh begins on a different frame for variety.
func newFysh(size int) (*canvas.Image, *fyne.Animation) {
	ico := &canvas.Image{}
	ico.Resize(fyne.NewSquareSize(float32(size)))
	offset := fyshCount % frameCount
	fyshCount++

	ico.Image = fyshes[offset]
	shown := offset
	anim := &fyne.Animation{
		Duration:    fyshFrameTime * frameCount,
		RepeatCount: fyne.AnimationRepeatForever,
		Curve:       fyne.AnimationLinear,
		Tick: func(done float32) {
			id := (offset + int(done*frameCount)) % frameCount
			if id == shown {
				return
			}

			shown = id
			ico.Image = fyshes[id]
			ico.Refresh()
		},
	}

	return ico, anim
}

func formattedTime(format string) string { // matching the desktop format
	if format == "12h" {
		return time.Now().Format("3:04pm")
	}

	return time.Now().Format("15:04")
}

// clockText keeps the label showing the current time until done is closed.
func clockText(t *canvas.Text, format string, done chan struct{}) {
	oldTime := ""
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-done:
			return
		case <-tick.C:
		}

		txt := formattedTime(format)
		if txt == oldTime {
			continue
		}

		oldTime = txt
		fyne.DoAndWait(func() {
			t.Text = txt
			t.Resize(t.MinSize())
		})
	}
}

func (s *ScreenSaver) startedInput(w fyne.Window) {
	if s.started.After(time.Now().Add(time.Millisecond * -200)) {
		return // something flickering as we start
	}

	// Input means we are plainly awake, even if the resume passed us by.
	if s.asleep.CompareAndSwap(true, false) {
		s.startUI()
	}

	if !s.Lock || (!s.LockImmediately && s.started.After(time.Now().Add(time.Second*-3))) {
		showCursor(w)
		s.unlock()
		return
	}

	err := showLogin(s.unlock, w)
	if err != nil {
		fyne.LogError("failed to prepare login screen", err)
		showCursor(w)
		s.unlock()
	}
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

// animate returns a looping animation that steps this layout's object along by
// one frame each time it ticks. If drawing stalls, because the machine suspended
// or an output powered down, the movement simply stops.
func (m *moveLayout) animate() *fyne.Animation {
	return &fyne.Animation{
		Duration:    animationPeriod,
		RepeatCount: fyne.AnimationRepeatForever,
		Curve:       fyne.AnimationLinear,
		Tick: func(float32) {
			if len(m.objs) == 0 || m.size.Width == 0 {
				return
			}

			m.move()
		},
	}
}

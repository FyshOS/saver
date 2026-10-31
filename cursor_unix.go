//go:build linux || openbsd || freebsd || netbsd || darwin

package saver

import (
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xfixes"
	"github.com/BurntSushi/xgb/xproto"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var (
	conn     *xgb.Conn
	connOnce sync.Once
)

// initCursor opens the X connection used to hide and show the cursor and checks
// It runs once for the process and publishes the connection only once it is ready.
func initCursor() {
	connOnce.Do(func() {
		c, err := xgb.NewConn()
		if err != nil {
			fyne.LogError("Failed to connect to X11 to hide cursor", err)
			return
		}

		err = xfixes.Init(c)
		if err != nil {
			fyne.LogError("Failed to init fixes extension to X11 to hide cursor", err)
			c.Close()
			return
		}

		r, err := xfixes.QueryVersion(c, 4, 0).Reply()
		if err != nil || r.MajorVersion < 4 {
			c.Close()
			return
		}
		conn = c
	})
}

func doHideCursor(ctx any) {
	switch win := ctx.(type) {
	case driver.X11WindowContext:
		if conn == nil {
			return
		}

		_ = xfixes.HideCursorChecked(conn, xproto.Window(win.WindowHandle)).Check()
	case driver.WaylandWindowContext:
		// TODO request cursor hide for Wayland
	}
}

func doShowCursor(ctx any) {
	switch win := ctx.(type) {
	case driver.X11WindowContext:
		if conn == nil {
			return
		}

		_ = xfixes.ShowCursorChecked(conn, xproto.Window(win.WindowHandle)).Check()
	case driver.WaylandWindowContext:
		// TODO request cursor hide for Wayland
	}
}

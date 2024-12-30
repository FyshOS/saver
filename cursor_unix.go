package main

import (
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xfixes"
	"github.com/BurntSushi/xgb/xproto"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

var conn *xgb.Conn

func doHideCursor(w fyne.Window, ctx any) {
	switch win := ctx.(type) {
	case driver.X11WindowContext:
		conn, _ = xgb.NewConn()
		//if err != nil {
		//	fyne.LogError("Failed to connect to X11 to hide cursor", err)
		//	return
		//}

		err := xfixes.Init(conn)
		if err != nil {
			fyne.LogError("Failed to init fixes extension to X11 to hide cursor", err)
			return
		}

		r, err := xfixes.QueryVersion(conn, 4, 0).Reply()
		if err != nil || r.MajorVersion < 4 {
			return
		}

		_ = xfixes.HideCursorChecked(conn, xproto.Window(win.WindowHandle)).Check()
	case driver.WaylandWindowContext:
		// TODO request cursor hide for Wayland
	}
}

func doShowCursor(w fyne.Window, ctx any) {
	switch win := ctx.(type) {
	case driver.X11WindowContext:

		_ = xfixes.ShowCursorChecked(conn, xproto.Window(win.WindowHandle)).Check()
	case driver.WaylandWindowContext:
		// TODO request cursor hide for Wayland
	}
}

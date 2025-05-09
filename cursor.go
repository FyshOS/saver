package saver

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func hideCursor(w fyne.Window) {
	if native, ok := w.(driver.NativeWindow); ok {
		native.RunNative(func(ctx any) {
			doHideCursor(ctx)
		})
	}
}

func showCursor(w fyne.Window) {
	if native, ok := w.(driver.NativeWindow); ok {
		native.RunNative(func(ctx any) {
			doShowCursor(ctx)
		})
	}
}

type cursorCapture struct {
	widget.BaseWidget

	moved func()
	since time.Time
}

func (c *cursorCapture) CreateRenderer() fyne.WidgetRenderer {
	c.since = time.Now()
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (c *cursorCapture) MouseIn(_ *desktop.MouseEvent) {

}

func (c *cursorCapture) MouseMoved(_ *desktop.MouseEvent) {
	if time.Now().Sub(c.since) < time.Second {
		return
	}

	c.moved()
}

func (c *cursorCapture) MouseOut() {

}

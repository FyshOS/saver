package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func hideCursor(w fyne.Window) {
	if native, ok := w.(driver.NativeWindow); ok {
		native.RunNative(func(ctx any) {
			doHideCursor(w, ctx)
		})
	}
}

func showCursor(w fyne.Window) {
	if native, ok := w.(driver.NativeWindow); ok {
		native.RunNative(func(ctx any) {
			doShowCursor(w, ctx)
		})
	}
}

type cursorCapture struct {
	widget.BaseWidget

	moved func()
}

func (c *cursorCapture) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (c *cursorCapture) MouseIn(ev *desktop.MouseEvent) {

}

func (c *cursorCapture) MouseMoved(ev *desktop.MouseEvent) {
	c.moved()
}

func (c *cursorCapture) MouseOut() {

}

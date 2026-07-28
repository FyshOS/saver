package saver

import (
	"fyne.io/fyne/v2"
)

// CustomUI is implemented by a screen saver's visual content. MakeUI builds the
// canvas content and DestroyUI tears it down when the window closes. Stop and
// Start pause and resume any animations.
type CustomUI interface {
	MakeUI(s *ScreenSaver) fyne.CanvasObject
	DestroyUI()
	Stop()
	Start()
}

func (s *ScreenSaver) SetCustomUI(ui CustomUI) {
	s.ui = ui
}

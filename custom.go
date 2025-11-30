package saver

import (
	"fyne.io/fyne/v2"
)

type CustomUI interface {
	MakeUI(s *ScreenSaver) fyne.CanvasObject
	DestroyUI()
}

func (s *ScreenSaver) SetCustomUI(ui CustomUI) {
	s.ui = ui
}

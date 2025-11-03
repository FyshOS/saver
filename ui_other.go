//go:build !linux && !openbsd && !freebsd && !netbsd && !darwin

package saver

import glfw "github.com/fyne-io/glfw-js"

func getMonitors() []*monitor {
	return []*monitor{&monitor{glfw.GetPrimaryMonitor()}}
}

type monitor struct {
	*glfw.Monitor
}

func (m *monitor) GetPos() (x, y int) {
	return 0, 0 // top left
}

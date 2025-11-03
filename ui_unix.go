//go:build linux || openbsd || freebsd || netbsd || darwin

package saver

import "github.com/go-gl/glfw/v3.3/glfw"

func getMonitors() []*glfw.Monitor {
	return glfw.GetMonitors()
}

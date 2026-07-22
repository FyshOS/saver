//go:build linux

package saver

import (
	"fyne.io/fyne/v2"

	"github.com/godbus/dbus/v5"
)

// watchSuspend subscribes to systemd-logind's PrepareForSleep signal and calls
// onChange(true) just before the system suspends and onChange(false) once it
// resumes. It returns a function that stops watching. If the system bus or
// logind is unavailable it logs the reason and returns a no-op stop, so callers
// need not special-case those environments.
func watchSuspend(onChange func(sleeping bool)) (stop func()) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fyne.LogError("screensaver: failed to connect to system bus for suspend events", err)
		return func() {}
	}

	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		dbus.WithMatchMember("PrepareForSleep"),
		dbus.WithMatchObjectPath("/org/freedesktop/login1"),
	); err != nil {
		fyne.LogError("screensaver: failed to watch for PrepareForSleep", err)
		_ = conn.Close()
		return func() {}
	}

	ch := make(chan *dbus.Signal, 4)
	conn.Signal(ch)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case sig, ok := <-ch:
				if !ok {
					return
				}
				if sig == nil || len(sig.Body) == 0 {
					continue
				}
				if sleeping, ok := sig.Body[0].(bool); ok {
					onChange(sleeping)
				}
			}
		}
	}()

	return func() {
		close(done)
		_ = conn.Close()
	}
}

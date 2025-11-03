//go:build !linux && !openbsd && !freebsd && !netbsd && !darwin

package saver

func initCursor() {
}

func doHideCursor(ctx any) {
}

func doShowCursor(ctx any) {
}

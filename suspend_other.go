//go:build !linux

package saver

// watchSuspend is a no-op on platforms without systemd-logind. See the linux
// implementation for details.
func watchSuspend(_ func(sleeping bool)) (stop func()) {
	return func() {}
}

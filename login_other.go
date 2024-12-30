//go:build !unix

package saver

func canUnlock(username, password string) bool {
	return true // kind of useless fallback but we don't have any way to check
}

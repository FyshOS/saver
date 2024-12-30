//go:build !unix

package main

func canUnlock(username, password string) bool {
	return true // kind of useless fallback but we don't have any way to check
}

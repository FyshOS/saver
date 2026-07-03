//go:build !unix || openbsd

package saver

// canUnlockFingerprint is unavailable without Linux PAM (pam_fprintd), so the
// fingerprint unlock loop simply never matches on these platforms.
func canUnlockFingerprint(username string) bool {
	return false
}

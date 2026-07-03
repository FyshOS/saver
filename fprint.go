package saver

import "time"

// startFingerprintUnlock runs fingerprint verification in the background while
// the unlock prompt is shown. Each attempt blocks in pam_fprintd until a finger
// is presented (or it times out); on a match, onMatch is invoked once and the
// loop stops. It is a no-op on machines without a sensor or without fingerprint
// login configured, where each attempt returns quickly and we idle between them.
//
// The returned function stops the loop. Because a verification attempt cannot be
// interrupted mid-scan, stopping is cooperative: an in-flight attempt finishes
// before the goroutine exits, and onMatch is never called after a stop.
func startFingerprintUnlock(username string, onMatch func()) (stop func()) {
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			if canUnlockFingerprint(username) {
				select {
				case <-done:
					return
				default:
					onMatch()
				}
				return
			}

			// Avoid a hot loop when no sensor / fingerprint auth is present.
			select {
			case <-done:
				return
			case <-time.After(time.Second):
			}
		}
	}()

	var stopped bool
	return func() {
		if stopped {
			return
		}
		stopped = true
		close(done)
	}
}

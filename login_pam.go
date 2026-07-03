//go:build unix && !openbsd

package saver

/*
#cgo LDFLAGS: -lpam
#cgo openbsd CFLAGS: -I/usr/local/include
#cgo openbsd LDFLAGS: -L/usr/local/lib
#cgo darwin CFLAGS: -I/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/usr/include

#include <stdbool.h>
#include <stdlib.h>
#include <unistd.h>

bool canUnlock(const char *username, const char *password);
bool canUnlockFingerprint(const char *username);
*/
import "C"

import "unsafe"

// canUnlock checks if the username/password combination is allowed.
func canUnlock(username, password string) bool {
	cUser := C.CString(username)
	cPass := C.CString(password)
	defer C.free(unsafe.Pointer(cUser))
	defer C.free(unsafe.Pointer(cPass))

	ret := bool(C.canUnlock(cUser, cPass))
	return ret
}

// canUnlockFingerprint blocks for one fingerprint verification attempt and
// reports whether a matching finger was presented.
func canUnlockFingerprint(username string) bool {
	cUser := C.CString(username)
	defer C.free(unsafe.Pointer(cUser))

	return bool(C.canUnlockFingerprint(cUser))
}

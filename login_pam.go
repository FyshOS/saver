//go:build unix

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
*/
import "C"

// canUnlock checks if the username/password combination is allowed.
func canUnlock(username, password string) bool {
	cUser := C.CString(username)
	cPass := C.CString(password)

	ret := bool(C.canUnlock(cUser, cPass))
	return ret
}

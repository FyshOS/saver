//go:build openbsd

package saver

/*
#include <stdbool.h>

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

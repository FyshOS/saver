//go:build openbsd

#include <stdbool.h>
#include <stdio.h>
#include <string.h>
#include <pwd.h>
#include <login_cap.h>
#include <bsd_auth.h>

bool canUnlock(const char *username, const char *password) {
    struct passwd *pwd;
    auth_session_t *as;
    login_cap_t *lc;
    char *class = NULL, *style = NULL;

    if ((as = auth_open()) == NULL)
        return false;

    if ((pwd = getpwnam(username)) == NULL)
        return false;

    if (!class && pwd && pwd->pw_class && pwd->pw_class[0] != '\0')
        class = strdup(pwd->pw_class);

    /* Get login class; if invalid style treat like unknown user. */
    lc = login_getclass(class);
    if (lc && (style = login_getstyle(lc, style, "auth-display_manager")) == NULL) {
        login_close(lc);
        return false;
    }
    login_close(lc);

    if (auth_userokay((char *)username, NULL, NULL, (char *)password) == 0)
	    return false;

    return true;
}

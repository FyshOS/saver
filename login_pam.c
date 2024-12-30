//go:build unix

// The code in this file is from the excellent blog post https://www.gulshansingh.com/posts/how-to-write-a-display-manager/

#include <security/pam_appl.h>
#if defined(__FreeBSD__) || defined(__OpenBSD__) || defined(__APPLE__)
#include <security/openpam.h>
#else
#include <security/pam_misc.h>
#endif

#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pwd.h>
#include <paths.h>
#include <unistd.h>
#include <sys/types.h>
#include <grp.h>

#define SERVICE_NAME "display_manager"

#define err(name)                                   \
    do {                                            \
        fprintf(stderr, "%s: %s\n", name,           \
                pam_strerror(pam_handle, result));  \
        end(result);                                \
        return false;                               \
    } while (1);                                    \

static pam_handle_t *pam_handle;

static void change_identity (struct passwd *pw) {
    if (initgroups(pw->pw_name, pw->pw_gid) == -1)
        _Exit(1);
    endgrent();
    if (setgid(pw->pw_gid) || setuid(pw->pw_uid))
        _Exit(1);
}


static int end(int last_result) {
    int result = pam_end(pam_handle, last_result);
    pam_handle = 0;
    return result;
}

static int conv(int num_msg, const struct pam_message **msg,
        struct pam_response **resp, void *appdata_ptr) {
    int i;

    *resp = calloc(num_msg, sizeof(struct pam_response));
    if (*resp == NULL) {
        return PAM_BUF_ERR;
    }

    int result = PAM_SUCCESS;
    for (i = 0; i < num_msg; i++) {
        char *username, *password;
        switch (msg[i]->msg_style) {
            case PAM_PROMPT_ECHO_ON:
                username = ((char **) appdata_ptr)[0];
                (*resp)[i].resp = strdup(username);
                break;
            case PAM_PROMPT_ECHO_OFF:
                password = ((char **) appdata_ptr)[1];
                (*resp)[i].resp = strdup(password);
                break;
            case PAM_ERROR_MSG:
                fprintf(stderr, "%s\n", msg[i]->msg);
                result = PAM_CONV_ERR;
                break;
            case PAM_TEXT_INFO:
                printf("%s\n", msg[i]->msg);
                break;
        }
        if (result != PAM_SUCCESS) {
            break;
        }
    }

    if (result != PAM_SUCCESS) {
        free(*resp);
        *resp = 0;
    }

    return result;
}

bool canUnlock(const char *username, const char *password) {
    const char *data[2] = {username, password};
    struct pam_conv pam_conv = {
        conv, data
    };

    int result = pam_start(SERVICE_NAME, username, &pam_conv, &pam_handle);
    if (result != PAM_SUCCESS) {
        err("pam_start");
    }

    result = pam_authenticate(pam_handle, 0);
    if (result != PAM_SUCCESS) {
        err("pam_authenticate");
        return false;
    }

    end(result);
    return true;
}

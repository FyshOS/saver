# If PREFIX isn't provided, we check for /usr/local and use that if it exists.
# Otherwise we fall back to using /usr.

LOCAL != test -d $(DESTDIR)/usr/local && echo -n "/local" || echo -n ""
LOCAL ?= $(shell test -d $(DESTDIR)/usr/local && echo "/local" || echo "")
PREFIX ?= /usr$(LOCAL)

build:
	go build ./cmd/fyshsaver

install:
	install -Dm00755 fyshsaver $(DESTDIR)$(PREFIX)/bin/fyshsaver

uninstall:
	-rm $(DESTDIR)$(PREFIX)/bin/fyshsaver

# === FyshOS packaging =========================================================
DEB_VERSION     ?=
DEB_NAME        ?= fyshsaver
DEB_SECTION     ?= x11
DEB_DESCRIPTION ?= FyshOS screen saver and locker
DEB_HOMEPAGE    ?= https://fyshos.com
DEB_SUDO        ?= -sudo
DEB_BUILD_DEPS  ?= libpam0g-dev libgl1-mesa-dev xorg-dev libwayland-dev \
                   libxkbcommon-dev

repo:
	fyshpkg make \
		-name "$(DEB_NAME)" \
		$(if $(DEB_VERSION),-version "$(DEB_VERSION)") \
		-section "$(DEB_SECTION)" \
		-description "$(DEB_DESCRIPTION)" \
		-homepage "$(DEB_HOMEPAGE)" \
		-build-deps "$(DEB_BUILD_DEPS)" \
		$(DEB_SUDO) $(FYSHPKG_FLAGS) \
		cmd/fyshsaver

.PHONY: build install uninstall repo

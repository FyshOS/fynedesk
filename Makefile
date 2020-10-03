# If PREFIX isn't provided, we check for /usr/local and use that if it exists.
# Otherwice we fall back to using /usr.

LOCAL != test -d /usr/local && echo -n "/local" || echo -n ""
LOCAL ?= $(shell test -d /usr/local && echo "/local" || echo "")
PREFIX ?= /usr$(LOCAL)

build:
	go build ./cmd/fynedesk_runner || (echo "Failed to build fynedesk_runner"; exit 1)
	go build ./cmd/fynedesk || (echo "Failed to build fynedesk"; exit 1)

install:
	install -Dm00755 fynedesk_runner $(DESTDIR)$(PREFIX)/bin/fynedesk_runner
	install -Dm00755 fynedesk $(DESTDIR)$(PREFIX)/bin/fynedesk
	install -Dm00644 fynedesk.desktop $(DESTDIR)$(PREFIX)/share/xsessions/fynedesk.desktop

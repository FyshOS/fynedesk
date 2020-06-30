# If PREFIX isn't provided, we check for /usr/local and use that if it exists. Otherwice we fall back to using /usr.
ifneq ("$(wildcard /usr/local)","")
PREFIX ?= /usr/local
else
PREFIX ?= /usr
endif

build:
	go build ./cmd/fynedesk_runner || (echo "Failed to build fynedesk_runner"; exit 1)
	go build ./cmd/fynedesk || (echo "Failed to build fynedesk"; exit 1)

install:
	install -Dm00755 fynedesk_runner $(DESTDIR)$(PREFIX)/bin/fynedesk_runner
	install -Dm00755 fynedesk $(DESTDIR)$(PREFIX)/bin/fynedesk
	install -Dm00644 fynedesk.desktop $(DESTDIR)$(PREFIX)/share/xsessions/fynedesk.desktop

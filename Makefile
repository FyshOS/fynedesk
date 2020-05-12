# If DESTDIR is not provided, we default to /usr/local if it exists and if it doesn't we fall back /usr.
ifneq ("$(wildcard /usr/local)","")
DESTDIR ?= /usr/local
else
DESTDIR ?= /usr
endif

build:
	go build ./cmd/fynedesk_runner || (echo "Failed to build fynedesk_runner"; exit 1)
	go build ./cmd/fynedesk || (echo "Failed to build fynedesk"; exit 1)

install:
	cp cmd/fynedesk_runner/fynedesk_runner $(DESTDIR)$(PREFIX)/bin
	cp cmd/fynedesk/fynedesk $(DESTDIR)$(PREFIX)/bin
	cp fynedesk.desktop $(DESTDIR)/share/xsessions


# If not DESTDIR is provided, we default to /usr/local if it exists and if it doesn't we fall back /usr.
ifneq ("$(wildcard /usr/local)","")
DESTDIR ?= /usr/local
else
DESTDIR ?= /usr
endif

build:
	go build ./cmd/fynedesk_runner
	go build ./cmd/fynedesk

install: build
	sudo cp cmd/fynedesk_runner/fynedesk_runner $(DESTDIR)$(PREFIX)/bin
	sudo cp cmd/fynedesk/fynedesk $(DESTDIR)$(PREFIX)/bin
	sudo cp fynedesk.desktop $(DESTDIR)/share/xsessions

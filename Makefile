build:
	go build ./cmd/fynedesk_runner
	go build ./cmd/fynedesk

install: build
	cp cmd/fynedesk_runner/fynedesk_runner $(DESTDIR)$(PREFIX)/bin
	cp cmd/fynedesk/fynedesk $(DESTDIR)$(PREFIX)/bin
	cp fynedesk.desktop $(DESTDIR)/share/xsessions

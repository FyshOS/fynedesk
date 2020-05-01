package wm

import (
	"sync"

	"fyne.io/fyne"
)

// ShortcutHandler is a simple implementation for tracking registered shortcuts
type ShortcutHandler struct {
	mu    sync.RWMutex
	entry map[fyne.Shortcut]func(fyne.Shortcut)
}

// TypedShortcut handle the registered shortcut
func (sh *ShortcutHandler) TypedShortcut(shortcut fyne.Shortcut) {
	var match func(fyne.Shortcut)
	for s, f := range sh.entry {
		if s.ShortcutName() == shortcut.ShortcutName() {
			match = f
		}
	}
	if match == nil {
		return
	}

	match(shortcut)
}

// AddShortcut register an handler to be executed when the shortcut action is triggered
func (sh *ShortcutHandler) AddShortcut(shortcut fyne.Shortcut, handler func(shortcut fyne.Shortcut)) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.entry == nil {
		sh.entry = make(map[fyne.Shortcut]func(fyne.Shortcut))
	}
	sh.entry[shortcut] = handler
}

// Shortcuts returns the list of all registered shortcuts
func (sh *ShortcutHandler) Shortcuts() []fyne.Shortcut {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	var shorts []fyne.Shortcut
	for s := range sh.entry {
		shorts = append(shorts, s)
	}
	return shorts
}

// ShortcutManager is an interface that we can use to check for the handler capabilities of a desktop
type ShortcutManager interface {
	Shortcuts() []fyne.Shortcut
	TypedShortcut(fyne.Shortcut)
}

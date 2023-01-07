package wm

import (
	"sync"

	"fyne.io/fyne/v2"

	"fyshos.com/fynedesk"
)

// ShortcutHandler is a simple implementation for tracking registered shortcuts
type ShortcutHandler struct {
	mu    sync.RWMutex
	entry map[*fynedesk.Shortcut]func()
}

// TypedShortcut handle the registered shortcut
func (sh *ShortcutHandler) TypedShortcut(shortcut fyne.Shortcut) {
	var matched func()
	for s, f := range sh.entry {
		if s.ShortcutName() == shortcut.ShortcutName() {
			matched = f
		}
	}
	if matched == nil {
		return
	}

	matched()
}

// AddShortcut register an handler to be executed when the shortcut action is triggered
func (sh *ShortcutHandler) AddShortcut(shortcut *fynedesk.Shortcut, handler func()) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.entry == nil {
		sh.entry = make(map[*fynedesk.Shortcut]func())
	}
	sh.entry[shortcut] = handler
}

// Shortcuts returns the list of all registered shortcuts
func (sh *ShortcutHandler) Shortcuts() []*fynedesk.Shortcut {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	var shorts []*fynedesk.Shortcut
	for s := range sh.entry {
		shorts = append(shorts, s)
	}
	return shorts
}

// ShortcutManager is an interface that we can use to check for the handler capabilities of a desktop
type ShortcutManager interface {
	Shortcuts() []*fynedesk.Shortcut
	TypedShortcut(fyne.Shortcut)
}

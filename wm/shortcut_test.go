package wm

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyshos.com/fynedesk"
	"github.com/stretchr/testify/assert"
)

func TestShortcutHandler_Shortcuts(t *testing.T) {
	m := &ShortcutHandler{}
	assert.Equal(t, 0, len(m.Shortcuts()))

	m.AddShortcut(fynedesk.NewShortcut("Hint", fyne.KeyB, fyne.KeyModifierSuper), func() {})
	assert.Equal(t, 1, len(m.Shortcuts()))
}

func TestShortcutHandler_TypedShortcut(t *testing.T) {
	m := &ShortcutHandler{}
	called := false
	key := fynedesk.NewShortcut("Hint", fyne.KeyH, fyne.KeyModifierSuper)
	m.AddShortcut(key, func() {
		called = true
	})
	m.TypedShortcut(key)
	assert.True(t, called)
}

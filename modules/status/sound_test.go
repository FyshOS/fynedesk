package status

import (
	"testing"

	"fyne.io/fyne/v2/test"

	wmtheme "fyshos.com/tyde/theme"

	"github.com/stretchr/testify/assert"
)

func TestStartsWith(t *testing.T) {
	assert.True(t, startsWith("vol", "volume"))
	assert.True(t, startsWith("volume", "volume"))
	assert.True(t, startsWith("volume up", "volume"))

	assert.False(t, startsWith("", "volume"))
	assert.False(t, startsWith("voo", "volume"))
}

func TestVolItem_Title(t *testing.T) {
	i := &volItem{input: "50"}
	assert.Equal(t, "Volume 50%", i.Title())

	i = &volItem{input: "mute"}
	assert.Equal(t, "Mute volume", i.Title())

	i = &volItem{input: "unmute"}
	assert.Equal(t, "Unmute volume", i.Title())

	i = &volItem{input: "up"}
	assert.Equal(t, "Volume up", i.Title())

	i = &volItem{input: "down"}
	assert.Equal(t, "Volume down", i.Title())

	i = &volItem{input: ""}
	assert.Equal(t, "", i.Title())
}

func TestVolItem_Icon(t *testing.T) {
	i := &volItem{input: "50"}
	assert.NotNil(t, i.Icon())
}

// The tray icon follows the mute state whichever path changed it, and putting
// it in step is safe before the status area exists - a command from tyde_ctl or
// the assistant can arrive at any time.
func TestSound_UpdateIcon(t *testing.T) {
	test.NewApp()
	b := &sound{}
	b.updateIcon(50, true) // no status widget yet

	b.mute = newScrollButton(wmtheme.SoundHighIcon)
	b.updateIcon(50, true)
	assert.Equal(t, wmtheme.MuteIcon, b.mute.Icon)

	b.updateIcon(10, false)
	assert.Equal(t, wmtheme.SoundLowIcon, b.mute.Icon)

	b.updateIcon(80, false)
	assert.Equal(t, wmtheme.SoundHighIcon, b.mute.Icon)
}

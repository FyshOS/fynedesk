package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestDeskBadgeShow(t *testing.T) {
	test.NewTempApp(t)
	b := newDeskBadge()
	b.Resize(fyne.NewSize(800, 600))
	assert.False(t, b.Visible())

	b.show(2)
	assert.True(t, b.Visible())
	assert.Equal(t, "3", b.text.Text, "desktops are numbered from one")
	assert.InDelta(t, 90, b.text.TextSize, 0.01, "sized from the screen height")
	_, _, _, a := b.text.Color.RGBA()
	assert.Equal(t, uint32(0xffff), a, "fully opaque while held")

	// The number is centred in its box, and the box centred on the layer.
	r := b.CreateRenderer()
	r.Layout(b.Size())
	bg := b.bg
	assert.InDelta(t, 400, bg.Position().X+bg.Size().Width/2, 1)
	assert.InDelta(t, 300, bg.Position().Y+bg.Size().Height/2, 1)
	assert.GreaterOrEqual(t, bg.Size().Width, bg.Size().Height)

	b.cancel()
	b.setOpacity(0)
	_, _, _, a = b.text.Color.RGBA()
	assert.Equal(t, uint32(0), a, "transparent once faded")
}

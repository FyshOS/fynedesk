package wm

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
)

func TestFindObjectAtPixelPositionMatching(t *testing.T) {
	l := widget.NewRichTextFromMarkdown("* Test")
	e := widget.NewEntry()
	w := test.NewWindow(
		container.NewGridWithColumns(1, l, e))

	assert.Equal(t, l, FindObjectAtPixelPositionMatching(8, 8, w.Canvas(), func(fyne.CanvasObject) bool {
		return true
	}))
	assert.Equal(t, e, FindObjectAtPixelPositionMatching(8, 52, w.Canvas(), func(o fyne.CanvasObject) bool {
		_, ok := o.(*widget.Entry)
		return ok
	}))
	assert.Nil(t, FindObjectAtPixelPositionMatching(68, 68, w.Canvas(), func(fyne.CanvasObject) bool {
		return true
	}))
	assert.Nil(t, FindObjectAtPixelPositionMatching(8, 8, w.Canvas(), func(fyne.CanvasObject) bool {
		return false
	}))
}

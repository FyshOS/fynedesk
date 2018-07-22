package desktop

import "testing"

import "github.com/fyne-io/fyne/test"

import "github.com/stretchr/testify/assert"

func TestNewDesktop(t *testing.T) {
	app := test.NewApp()
	desktop := NewDesktop(app)

	min := desktop.Content().MinSize()
	assert.True(t, min.Width > 300)
	assert.True(t, min.Height > 200)
}

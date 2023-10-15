package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyshos.com/fynedesk"
	"fyshos.com/fynedesk/test"
)

func testWindows() []fynedesk.Window {
	desk := &desktop{}
	desk.settings = test.NewSettings()
	fynedesk.SetInstance(desk)
	return []fynedesk.Window{
		test.NewWindow("App1"),
		test.NewWindow("App2"),
		test.NewWindow("App3"),
	}
}

func TestShowAppSwitcher(t *testing.T) {
	wins := testWindows()
	s := NewAppSwitcher(wins, test.NewAppProvider())
	s.Show()

	assert.NotNil(t, s.win)
	assert.Equal(t, 1, s.currentIndex())
}

func TestShowAppSwitcherReverse(t *testing.T) {
	wins := testWindows()
	s := NewAppSwitcherReverse(wins, test.NewAppProvider())
	s.Show()

	assert.NotNil(t, s.win)
	assert.Equal(t, len(wins)-1, s.currentIndex())
}

func TestSwitcher_Next(t *testing.T) {
	wins := testWindows()
	s := NewAppSwitcher(wins, test.NewAppProvider())

	current := s.currentIndex()
	s.Next()
	assert.Equal(t, current+1, s.currentIndex())

	s.setCurrent(len(s.icons) - 1)
	s.Next()
	assert.Equal(t, 0, s.currentIndex())
}

func TestSwitcher_Previous(t *testing.T) {
	wins := testWindows()
	s := NewAppSwitcher(wins, test.NewAppProvider())

	current := s.currentIndex()
	s.Previous()
	assert.Equal(t, current-1, s.currentIndex())

	s.setCurrent(0)
	s.Previous()
	assert.Equal(t, len(s.icons)-1, s.currentIndex())
}

func TestSwitcher_HideApply(t *testing.T) {
	wins := testWindows()
	s := NewAppSwitcher(wins, test.NewAppProvider())
	s.Show()

	s.HideApply()
	assert.True(t, wins[s.currentIndex()].(*test.Window).TopWindow())
}

func TestSwitcher_HideCancel(t *testing.T) {
	wins := testWindows()
	s := NewAppSwitcher(wins, test.NewAppProvider())
	s.Show()

	s.HideCancel()
	assert.False(t, wins[s.currentIndex()].(*test.Window).TopWindow())
}

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fynedesk"
)

func testWindows() []fynedesk.Window {
	desk := &deskLayout{}
	desk.settings = &testSettings{}
	fynedesk.SetInstance(desk)
	return []fynedesk.Window{
		&dummyWindow{name: "App1"},
		&dummyWindow{name: "App2"},
		&dummyWindow{name: "App3"},
	}
}

func TestShowAppSwitcher(t *testing.T) {
	wins := testWindows()
	s := ShowAppSwitcher(wins, &testAppProvider{})

	assert.NotNil(t, s.win)
	assert.Equal(t, 1, s.currentIndex())
}

func TestShowAppSwitcherReverse(t *testing.T) {
	wins := testWindows()
	s := ShowAppSwitcherReverse(wins, &testAppProvider{})

	assert.NotNil(t, s.win)
	assert.Equal(t, len(wins)-1, s.currentIndex())
}

func TestSwitcher_Next(t *testing.T) {
	wins := testWindows()
	s := ShowAppSwitcher(wins, &testAppProvider{})

	current := s.currentIndex()
	s.Next()
	assert.Equal(t, current+1, s.currentIndex())

	s.setCurrent(len(s.icons) - 1)
	s.Next()
	assert.Equal(t, 0, s.currentIndex())
}

func TestSwitcher_Previous(t *testing.T) {
	wins := testWindows()
	s := ShowAppSwitcher(wins, &testAppProvider{})

	current := s.currentIndex()
	s.Previous()
	assert.Equal(t, current-1, s.currentIndex())

	s.setCurrent(0)
	s.Previous()
	assert.Equal(t, len(s.icons)-1, s.currentIndex())
}

func TestSwitcher_HideApply(t *testing.T) {
	wins := testWindows()
	s := ShowAppSwitcher(wins, &testAppProvider{})

	s.HideApply()
	assert.True(t, wins[s.currentIndex()].(*dummyWindow).raised)
}

func TestSwitcher_HideCancel(t *testing.T) {
	wins := testWindows()
	s := ShowAppSwitcher(wins, &testAppProvider{})

	s.HideCancel()
	assert.False(t, wins[s.currentIndex()].(*dummyWindow).raised)
}

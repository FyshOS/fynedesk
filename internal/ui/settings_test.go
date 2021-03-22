package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fynedesk/test"
)

func TestDeskSettings_IsModuleEnabled(t *testing.T) {
	s := test.NewSettings()
	s.SetModuleNames([]string{"Yes", "maybe"})

	assert.True(t, isModuleEnabled("Yes", s))
	assert.True(t, isModuleEnabled("maybe", s))
	assert.False(t, isModuleEnabled("Maybe", s))
	assert.False(t, isModuleEnabled("No", s))
}

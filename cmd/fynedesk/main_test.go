package main

import (
	"testing"

	"fyne.io/fyne/test"

	"github.com/fyne-io/desktop"
)

func TestNewDesktop(t *testing.T) {
	app := test.NewApp()
	desk := desktop.NewEmbeddedDesktop(app)

	desk.Run()
}

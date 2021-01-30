package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestNewDesktop(t *testing.T) {
	app := test.NewApp()
	desk := setupDesktop(app)

	desk.Run()
}

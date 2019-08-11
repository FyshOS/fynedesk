package main

import (
	"testing"

	"fyne.io/fyne/test"
)

func TestNewDesktop(t *testing.T) {
	app := test.NewApp()
	desk := setupDesktop(app)

	desk.Run()
}

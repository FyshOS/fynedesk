package desktop

import "fyne.io/fyne"

// ModuleMetadata is the information required to describe a module in FyneDesk
type ModuleMetadata struct {
	Name string
}

// Module marks the required methods of a pluggable module in FyneDesk.
type Module interface {
	Metadata() ModuleMetadata
}

// StatusAreaModule describes a module that can add items to the status area
// (the bottom of the widget panel)
type StatusAreaModule interface {
	Module
	StatusAreaWidget() fyne.CanvasObject
}

// ScreenAreaModule describes a module that can draw on the whole screen -
// these items will appear over the background image.
type ScreenAreaModule interface {
	Module
	ScreenAreaWidget() fyne.CanvasObject
}

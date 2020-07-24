package fynedesk

import "fyne.io/fyne"

// ModuleMetadata is the information required to describe a module in FyneDesk
type ModuleMetadata struct {
	Name        string
	NewInstance func() Module
}

// KeyBindModule marks a module that provides key bindings.
// This is optional but can be enabled for any module by implementing the interface.
type KeyBindModule interface {
	Shortcuts() map[*Shortcut]func()
}

// Module marks the required methods of a pluggable module in FyneDesk.
type Module interface {
	Metadata() ModuleMetadata
	Destroy()
}

// LaunchSuggestion represents an item that can appear in the app launcher and be actioned on tap
type LaunchSuggestion interface {
	Icon() fyne.Resource
	Title() string
	Launch()
}

// LaunchSuggestionModule is a module that can provide suggestions for the app launcher
type LaunchSuggestionModule interface {
	Module
	LaunchSuggestions(string) []LaunchSuggestion
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

var modules []ModuleMetadata

// AvailableModules lists all of the FyneDesk modules that were found at runtime
func AvailableModules() []ModuleMetadata {
	return modules
}

// RegisterModule adds a module to the list of available modules.
// New module packages should probably call this in their init().
func RegisterModule(m ModuleMetadata) {
	modules = append(modules, m)
}

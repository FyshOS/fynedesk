package tyde

import "fyne.io/fyne/v2"

// ModuleMetadata is the information required to describe a module in Tyde.
type ModuleMetadata struct {
	Name        string
	NewInstance func() Module
}

// KeyBindModule marks a module that provides key bindings.
// This is optional but can be enabled for any module by implementing the interface.
type KeyBindModule interface {
	Shortcuts() map[*Shortcut]func()
}

// Module marks the required methods of a pluggable module in Tyde.
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

// Command is an action that can be asked for by name from outside the desktop UI.
type Command struct {
	Name        string   // the word that invokes it, e.g. "brightness"
	Description string   // one line saying what running it does
	Args        string   // what may follow the name, empty when it takes none
	Examples    []string // optional complete invocations, e.g. "brightness 50"

	// Run performs the command with any parameters. Refuse arguments you cannot act on
	// with an error rather than doing nothing quietly.
	Run func(args string) (string, error)
}

// CommandModule is a module that offers commands that can be executed. Implementing this is
// all that is needed for a module's actions to appear in tyde_ctl and the AI assistant.
type CommandModule interface {
	Module
	Commands() []Command
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

// OverlayAreaModule describes a module that can draw over the whole screen
// ABOVE application windows (but below the bar, widget panel, menus and the
// mouse cursor). The returned content must be visual-only (non-interactive)
// so that it does not capture pointer input destined for the windows beneath
// it - decorative overlays such as desktop pets are the intended use.
type OverlayAreaModule interface {
	Module
	OverlayAreaWidget() fyne.CanvasObject
}

// WindowAccessory is a decorative canvas object drawn with a particular window:
// at the object's own position relative to that window's top left corner, and at
// the window's z-level so windows stacked above it occlude the decoration.
type WindowAccessory struct {
	Object fyne.CanvasObject
	Window Window
}

// WindowAccessoryModule contributes WindowAccessory items that the compositor
// interleaves among the window images - e.g. a desktop pet that walks along a
// window's title bar and disappears behind windows above it. Call
// Desktop.RefreshWindowAccessories when the set, stacking or offsets of the
// accessories change so the compositor re-assembles them; following a window as
// it moves is handled for you.
type WindowAccessoryModule interface {
	Module
	WindowAccessories() []WindowAccessory
}

var modules []ModuleMetadata

// AvailableModules lists all the Tyde modules that were found at runtime.
func AvailableModules() []ModuleMetadata {
	return modules
}

// RegisterModule adds a module to the list of available modules.
// New module packages should probably call this in their init().
func RegisterModule(m ModuleMetadata) {
	modules = append(modules, m)
}

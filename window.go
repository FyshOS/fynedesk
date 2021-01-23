package fynedesk

import (
	"image"

	"fyne.io/fyne/v2"
)

// Window represents a single managed window within a window manager.
// There may be borders or not depending on configuration.
type Window interface {
	Focused() bool      // Is this the currently focused window?
	Fullscreened() bool // Is the window Fullscreen?
	Iconic() bool       // Is the window Iconified?
	Maximized() bool    // Is the window Maximized?
	TopWindow() bool    // Is this the window on top?

	Capture() image.Image // Capture the contents of this window to an image
	Close()               // Close this window and possibly the application running it
	Focus()               // Ask this window to get input focus
	Fullscreen()          // Request to fullscreen this window
	Iconify()             // Request to iconify this window
	Maximize()            // Request to resize this window to it's largest possible size
	RaiseAbove(Window)    // Raise this window above a given other window
	RaiseToTop()          // Raise this window to the top of the stack
	Unfullscreen()        // Request to unfullscreen this window
	Uniconify()           // Request to restore this window and possibly children of this window from being minimized
	Unmaximize()          // Request to restore this window to its size before being maximized

	Properties() WindowProperties // Request the properties set on this window
}

// WindowProperties encapsulates the metadata that a window can provide.
type WindowProperties interface {
	Class() []string     // The class of this window
	Command() string     // The command of this window
	Decorated() bool     // Should this window have borders drawn?
	Icon() fyne.Resource // The icon of this window
	IconName() string    // The icon name of this window
	SkipTaskbar() bool   // Should this window be added to the taskbar?
	Title() string       // The name of this window
}

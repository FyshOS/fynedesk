package fynedesk

import "fyne.io/fyne"

// WindowManager describes a full window manager which may be loaded as part of the setup.
type WindowManager interface {
	Stack
	AddStackListener(StackListener)

	Blank()
	Close()
}

// Window represents a single managed window within a window manager.
// There may be borders or not depending on configuration.
type Window interface {
	Class() []string     // The class of this window
	Command() string     // The command of this window
	Decorated() bool     // Should this window have borders drawn?
	Focused() bool       // Is this the currently focused window?
	Fullscreened() bool  // Is the window Fullscreen?
	Icon() fyne.Resource // The icon of this window
	Iconic() bool        // Is the window Iconified?
	IconName() string    // The icon name of this window
	Maximized() bool     // Is the window Maximized?
	SkipTaskbar() bool   // Should this window be added to the taskbar?
	Title() string       // The name of this window
	TopWindow() bool     // Is this the window on top?

	Close()            // Close this window and possibly the application running it
	Focus()            // Ask this window to get input focus
	Fullscreen()       // Request to fullscreen this window
	Iconify()          // Request to iconify this window
	Maximize()         // Request to resize this window to it's largest possible size
	RaiseAbove(Window) // Raise this window above a given other window
	RaiseToTop()       // Raise this window to the top of the stack
	Unfullscreen()     // Request to unfullscreen this window
	Uniconify()        // Request to restore this window and possibly children of this window from being minimized
	Unmaximize()       // Request to restore this window to its size before being maximized
}

// Stack describes an ordered list of windows.
// The order of the windows in this list matches the stacking order on screen.
// TopWindow() returns the 0th element with each item after that being stacked below the previous.
type Stack interface {
	AddWindow(Window)    // Add a new window to the stack
	RaiseToTop(Window)   // Request that the passed window become top of the stack.
	RemoveWindow(Window) // Remove a specified window from the stack
	TopWindow() Window   // Get the currently top most window
	Windows() []Window   // Return a list of all managed windows. This should not be modified
}

// StackListener is used to listen for events in the window manager stack (window list).
// See WindowManager.AddStackListener.
type StackListener interface {
	WindowAdded(Window)
	WindowRemoved(Window)
}

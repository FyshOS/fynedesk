package fynedesk

import "fyne.io/fyne"

// WindowManager describes a full window manager which may be loaded as part of the setup.
type WindowManager interface {
	Stack
	AddStackListener(StackListener)

	Close()
	SetRoot(fyne.Window)
	Blank()
}

// Window represents a single managed window within a window manager.
// There may be borders or not depending on configuration.
type Window interface {
	Decorated() bool     // Should this window have borders drawn?
	Title() string       // The name of this window
	Class() []string     // The class of this window
	Command() string     // The command of this window
	IconName() string    // The icon name of this window
	Icon() fyne.Resource // The icon of this window
	Fullscreened() bool  // Is the window Fullscreen?
	Iconic() bool        // Is the window Iconified?
	Maximized() bool     // Is the window Maximized?
	TopWindow() bool     // Is this the window on top?
	SkipTaskbar() bool   // Should this window be added to the taskbar?
	Focused() bool       // Is this the currently focused window?

	Focus()            // Ask this window to get input focus
	Close()            // Close this window and possibly the application running it
	Fullscreen()       // Request to fullscreen this window
	Unfullscreen()     // Request to unfullscreen this window
	Iconify()          // Request to inimize this window and possibly children of this window
	Uniconify()        // Request to restore this window and possibly children of this window from being minimized
	Maximize()         // Request to resize this window to it's largest possible size
	Unmaximize()       // Request to restore this window to its size before being maximized
	RaiseAbove(Window) // Raise this window above a given other window
	RaiseToTop()       // Raise this window to the top of the stack
}

// Stack describes an ordered list of windows.
// The order of the windows in this list matches the stacking order on screen.
// TopWindow() returns the 0th element with each item after that being stacked below the previous.
type Stack interface {
	AddWindow(Window)    // Add a new window to the stack
	RemoveWindow(Window) // Remove a specified window from the stack

	TopWindow() Window // Get the currently top most window
	Windows() []Window // Return a list of all managed windows. This should not be modified

	RaiseToTop(Window) // Request that the passed window become top of the stack.
}

// StackListener is used to listen for events in the window manager stack (window list).
// See WindowManager.AddStackListener.
type StackListener interface {
	WindowAdded(Window)
	WindowRemoved(Window)
}

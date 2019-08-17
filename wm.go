package desktop

// WindowManager describes a full window manager which may be loaded as part of the setup.
type WindowManager interface {
	Stack
	AddStackListener(StackListener)

	Close()
}

// Window represents a single managed window within a window manager.
// There may be borders or not depending on configuration.
type Window interface {
	Decorated() bool  // Should this window have borders drawn?
	Title() string    // The name of this window
	Class() []string  // The class of this window
	Command() string  // The command of this window
	IconName() string // The icon name of this window

	Focus()            // Ask this window to get input focus
	Close()            // Close this window and possibly the application running it
	RaiseAbove(Window) // Raise this window above a given other window
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

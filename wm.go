package fynedesk

import "image"

// WindowManager describes a full window manager which may be loaded as part of the setup.
type WindowManager interface {
	Stack
	AddStackListener(StackListener)

	Blank()
	Capture() image.Image // Capture the contents of the whole desktop to an image
	Close()
	Run()
}

// Stack describes an ordered list of windows.
// The order of the windows in this list is bottom to top from what is visible on screen.
// TopWindow() returns the last element in the stack which is the top most window on screen.
type Stack interface {
	AddWindow(Window)    // Add a new window to the stack
	RaiseToTop(Window)   // Request that the passed window become top of the stack
	RemoveWindow(Window) // Remove a specified window from the stack
	TopWindow() Window   // Get the top most window currently visible
	Windows() []Window   // Return a list of all managed windows. This should not be modified
}

// StackListener is used to listen for events in the window manager stack (window list).
// See WindowManager.AddStackListener.
type StackListener interface {
	WindowAdded(Window)
	WindowRemoved(Window)
}

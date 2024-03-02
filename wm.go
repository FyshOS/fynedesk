package fynedesk

import (
	"image"

	"fyne.io/fyne/v2"
)

// WindowManager describes a full window manager which may be loaded as part of the setup.
type WindowManager interface {
	Stack
	AddStackListener(StackListener)

	Blank()
	Capture() image.Image // Capture the contents of the whole desktop to an image
	Close()
	Run()
	ShowOverlay(fyne.Window, fyne.Size, fyne.Position)
	ShowMenuOverlay(*fyne.Menu, fyne.Size, fyne.Position)
	ShowModal(fyne.Window, fyne.Size)
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

	WindowMoved(Window)
	WindowOrderChanged()
}

package test

import "fyne.io/fynedesk"

// Window is an in-memory virtual window for test purposes
type Window struct {
	props dummyProperties

	iconic, focused, fullscreen, maximized, raised bool
}

// NewWindow creates a virtual window with the given title ("" is acceptable)
func NewWindow(title string) *Window {
	win := &Window{}
	win.props.name = title
	return win
}

// Close this test window
func (w *Window) Close() {
	// no-op
}

// Fullscreened returns true if this window has been made full screen
func (w *Window) Fullscreened() bool {
	return w.fullscreen
}

// Focused returns true if this window has requested focus
func (w *Window) Focused() bool {
	return w.focused
}

// Focus sets this window to have input focus
func (w *Window) Focus() {
	w.focused = true
}

// Fullscreen simulates this window becoming full screen
func (w *Window) Fullscreen() {
	w.fullscreen = true
}

// Iconic returns true if this window has been iconified
func (w *Window) Iconic() bool {
	return w.iconic
}

// Iconify sets this window to be reduced to an icon
func (w *Window) Iconify() {
	w.iconic = true
}

// Maximize simulates this window becoming maximized
func (w *Window) Maximize() {
	w.maximized = true
}

// Maximized returns true if this window has been made maximized
func (w *Window) Maximized() bool {
	return false
}

// Properties obtains the window properties currently set
func (w *Window) Properties() fynedesk.WindowProperties {
	return w.props
}

// RaiseAbove sets this window to be above the passed window
func (w *Window) RaiseAbove(fynedesk.Window) {
	// no-op (this is instructing the window after stack changes)
}

// RaiseToTop sets this window to be the topmost
func (w *Window) RaiseToTop() {
	w.raised = true
}

// SetClass is a test utility to set the class property of this window
func (w *Window) SetClass(class []string) {
	w.props.class = class
}

// SetCommand is a test utility to set the command property of this window
func (w *Window) SetCommand(cmd string) {
	w.props.cmd = cmd
}

// SetIconName is a test utility to set the icon-name property of this window
func (w *Window) SetIconName(name string) {
	w.props.iconName = name
}

// TopWindow returns true if this window has been raised above all others
func (w *Window) TopWindow() bool {
	return w.raised
}

// Unfullscreen removes the fullscreen state of this window
func (w *Window) Unfullscreen() {
	w.fullscreen = false
}

// Uniconify returns this window to its normal state
func (w *Window) Uniconify() {
	w.iconic = false
}

// Unmaximize removes the maximized state of this window
func (w *Window) Unmaximize() {
	w.maximized = false
}

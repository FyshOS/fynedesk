/*
Package keybind provides an easy to use interface to assign callback functions
to human readable key sequences.

Working with the X keyboard encoding is not an easy task, and the keybind
package attempts to encapsulate much of the complexity. Namely, the keybind
package exports two function types: KeyPressFun and KeyReleaseFun. Values of
these types are functions, and have a method called 'Connect' that attaches
an event handler to be run when a particular key press is issued.

This is virtually identical to the way calbacks are attached using the xevent
package, but the Connect method in the keybind package has a couple extra
parameters that are specific to key bindings. Namely, the key sequence to
respond to (which is a combination of zero or more modifiers and exactly one
key) and whether to establish a passive grab. One can still attach callbacks
to Key{Press,Release} events using xevent, but it will be run for *all*
Key{Press,Release} events. (This is typically what one might do when setting up
an active grab.)

Initialization

Before using the keybind package, you should *always* make a single call to
keybind.Initialize for each X connection you're working with.

Key sequence format

Key sequences are human readable strings made up of zero or more modifiers and
exactly one key. Namely:

	[Mod[-Mod[...]]-]KEY

Where 'Mod' can be one of: shift, lock, control, mod1, mod2, mod3, mod4, mod5,
or any. You can view which keys activate each modifier using the 'xmodmap'
program. (If you don't have 'xmodmap', you could also run the 'xmodmap' example
in the examples directory.)

KEY must correspond to a valid keysym. Keysyms can be found by pressing keys
using the 'xev' program. Alternatively, you may inspect the 'keysyms' map in
xgbutil/keybind/keysymdef.go.

An example key sequence might look like 'Mod4-Control-Shift-t'. The keybinding
for that key sequence is activated when all three modifiers---mod4, control and
shift---are pressed along with the 't' key.

When to issue a passive grab

One of the parameters of the 'Connect' method is whether to issue a passive
grab or not. A passive grab is useful when you need to respond to a key press
on some parent window (like the root window) without actually focusing that
window. Not using a passive grab is useful when you only need to read key
presses when the window is focused.

For more information on the semantics of passive grabs, please see
http://tronche.com/gui/x/xlib/input/XGrabKey.html.

Also, by default, when issuing a grab on a particular (modifiers, keycode)
tuple, several grabs are actually made. In particular, for each grab requested,
another grab is made with the "num lock" mask, another grab is made with the
"caps lock" mask, and another grab is made with both the "num lock" and "caps
locks" masks. This allows key events to be reported regardless of whether
caps lock or num lock is enabled.

The extra masks added can be modified by changing the xevent.IgnoreMods slice.
If you modify xevent.IgnoreMods, it should be modified once on program startup
(i.e., before any key or mouse bindings are established) and never modified
again.

Key bindings on the root window example

To run a particular function whenever the 'Mod4-Control-Shift-t' key
combination is pressed (mod4 is typically the 'super' or 'windows' key, but can
vary based on your system), use something like:

	keybind.Initialize(XUtilValue) // call once before using keybind package
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			// do something when key is pressed
		}).Connect(XUtilValue, XUtilValue.RootWin(),
			"Mod4-Control-Shift-t", true)

Note that we issue a passive grab because Key{Press,Release} events on the root
window will only be reported when the root window has focus if no grab exists.

Key bindings on a window you create example

This code snippet attaches an event handler to some window you've created
without using a grab. Thus, the function will only be activated when the key
sequence is pressed and your window has focus.

	keybind.Initialize(XUtilValue) // call once before using keybind package
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			// do something when key is pressed
		}).Connect(XUtilValue, your-window-id, "Mod4-t", false)

Run a function on all key press events example

This code snippet actually does *not* use the keybind package, but illustrates
how the Key{Press,Release} event handlers in the xevent package can still be
useful. Namely, the keybind package discriminates among events depending upon
the key sequences pressed, whereas the xevent package is more general: it can
only discriminate at the event level.

	xevent.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			// do something when any key is pressed
		}).Connect(XUtilValue, your-window-id)

This is the kind of handler you might use to capture all key press events.
(i.e., if you have a text box for a user to type in.) Additionally, if you're
using this sort of event handler, keybind.LookupString will probably be of some
use. Its contract is that given a (modifiers, keycode) tuple
(information found in all Key{Press,Release} events) it will return a string
representation of the key pressed. We can modify the above example slightly to
echo the key pressed:

	xevent.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			fmt.Println("Key pressed:",
				keybind.LookupString(X, ev.State, ev.Detail))
		}).Connect(XUtilValue, your-window-id)

More examples

Complete working examples can be found in the examples directory of xgbutil. Of
particular interest are probably 'keypress-english' and 'simple-keybinding'.

*/
package keybind

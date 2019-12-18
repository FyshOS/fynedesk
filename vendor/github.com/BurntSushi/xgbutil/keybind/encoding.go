package keybind

/*
This file contains the logic to implement X's Keyboard Encoding
described here: http://goo.gl/qum9q

Essentially, LookupString is analogous to Xlib's XLookupString. It's useful
in determining the english string representation of modifiers + keycode.

It is not for the faint of heart.
*/

import (
	"strings"
	"unicode"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
)

// LookupString attempts to convert a (modifiers, keycode) to an english string.
// It essentially implements the rules described at http://goo.gl/qum9q
// Namely, the bulleted list that describes how key syms should be interpreted
// when various modifiers are pressed.
// Note that we ignore the logic that asks us to check if particular key codes
// are mapped to particular modifiers (i.e., "XK_Caps_Lock" to "Lock" modifier).
// We just check if the modifiers are activated. That's good enough for me.
// XXX: We ignore num lock stuff.
// XXX: We ignore MODE SWITCH stuff. (i.e., we don't use group 2 key syms.)
func LookupString(xu *xgbutil.XUtil, mods uint16,
	keycode xproto.Keycode) string {

	k1, k2, _, _ := interpretSymList(xu, keycode)

	shift := mods&xproto.ModMaskShift > 0
	lock := mods&xproto.ModMaskLock > 0
	switch {
	case !shift && !lock:
		return k1
	case !shift && lock:
		if len(k1) == 1 && unicode.IsLower(rune(k1[0])) {
			return k2
		} else {
			return k1
		}
	case shift && lock:
		if len(k2) == 1 && unicode.IsLower(rune(k2[0])) {
			return string(unicode.ToUpper(rune(k2[0])))
		} else {
			return k2
		}
	case shift:
		return k2
	}

	return ""
}

// ModifierString takes in a keyboard state and returns a string of all
// modifiers in the state.
func ModifierString(mods uint16) string {
	modStrs := make([]string, 0, 3)
	for i, mod := range Modifiers {
		if mod&mods > 0 && len(NiceModifiers[i]) > 0 {
			modStrs = append(modStrs, NiceModifiers[i])
		}
	}
	return strings.Join(modStrs, "-")
}

// KeyMatch returns true if a string representation of a key can
// be matched (case insensitive) to the (modifiers, keycode) tuple provided.
// String representations can be found in keybind/keysymdef.go
func KeyMatch(xu *xgbutil.XUtil,
	keyStr string, mods uint16, keycode xproto.Keycode) bool {

	guess := LookupString(xu, mods, keycode)
	return strings.ToLower(guess) == strings.ToLower(keyStr)
}

// interpretSymList interprets the keysym list for a particular keycode as
// described in the third and fourth paragraphs of http://goo.gl/qum9q
func interpretSymList(xu *xgbutil.XUtil, keycode xproto.Keycode) (
	k1 string, k2 string, k3 string, k4 string) {

	ks1 := KeysymGet(xu, keycode, 0)
	ks2 := KeysymGet(xu, keycode, 1)
	ks3 := KeysymGet(xu, keycode, 2)
	ks4 := KeysymGet(xu, keycode, 3)

	// follow the rules, third paragraph
	switch {
	case ks2 == 0 && ks3 == 0 && ks4 == 0:
		ks3 = ks1
	case ks3 == 0 && ks4 == 0:
		ks3 = ks1
		ks4 = ks2
	case ks4 == 0:
		ks4 = 0
	}

	// Now convert keysyms to strings, so we can do alphabetic shit.
	k1 = KeysymToStr(ks1)
	k2 = KeysymToStr(ks2)
	k3 = KeysymToStr(ks3)
	k4 = KeysymToStr(ks4)

	// follow the rules, fourth paragraph
	if k2 == "" {
		if len(k1) == 1 && unicode.IsLetter(rune(k1[0])) {
			k1 = string(unicode.ToLower(rune(k1[0])))
			k2 = string(unicode.ToUpper(rune(k1[0])))
		} else {
			k2 = k1
		}
	}
	if k4 == "" {
		if len(k3) == 1 && unicode.IsLetter(rune(k3[0])) {
			k3 = string(unicode.ToLower(rune(k3[0])))
			k4 = string(unicode.ToUpper(rune(k4[0])))
		} else {
			k4 = k3
		}
	}

	return
}

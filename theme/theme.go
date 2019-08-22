package theme // import "fyne.io/desktop/theme"

import "fyne.io/fyne/theme"

var (
	// PointerDefault is the standard pointer resource
	PointerDefault = pointerDefault

	// Background is the default background image
	Background = lochFynePicture

	// BatteryIcon is the material design icon for battery in light and dark theme
	BatteryIcon = theme.NewThemedResource(batteryIcon, nil)
	// BrightnessIcon is the material design icon for brightness in light and dark theme
	BrightnessIcon = theme.NewThemedResource(brightnessIcon, nil)
	// BrokenImageIcon is the material design icon for a broken image
	BrokenImageIcon = theme.NewThemedResource(brokenImageIcon, nil)

	// BorderWidth is the width of window frames
	BorderWidth = 4
	// ButtonWidth is the width of window buttons
	ButtonWidth = 28
	// TitleHeight is the height of a frame titleBar
	TitleHeight = 24
)

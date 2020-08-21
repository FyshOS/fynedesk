package theme // import "fyne.io/fynedesk/theme"

import (
	"image/color"

	"fyne.io/fyne/theme"
)

var (
	// PointerDefault is the standard pointer resource
	PointerDefault = pointerDefault

	// Background is the default background image
	Background = lochFynePicture
	// FyneAboutBackground is the image used as a background to the about screen
	FyneAboutBackground = fyneAboutBackground

	// BatteryIcon is the material design icon for battery in light and dark theme
	BatteryIcon = theme.NewThemedResource(batteryIcon, nil)
	// BrightnessIcon is the material design icon for brightness in light and dark theme
	BrightnessIcon = theme.NewThemedResource(brightnessIcon, nil)
	// CalculateIcon is the material design icon for a calculator in light and dark theme
	CalculateIcon = theme.NewThemedResource(calculateIcon, nil)
	// DisplayIcon is the material design icon for computer displays in light and dark theme
	DisplayIcon = theme.NewThemedResource(displayIcon, nil)
	// InternetIcon is the material design icon for the internet in light and dark theme
	InternetIcon = theme.NewThemedResource(internetIcon, nil)
	// UserIcon is the material design icon for a user in light and dark theme
	UserIcon = theme.NewThemedResource(personIcon, nil)

	// BrokenImageIcon is the material design icon for a broken image
	BrokenImageIcon = theme.NewThemedResource(brokenImageIcon, nil)
	// MaximizeIcon is the material design icon for maximizing a window
	MaximizeIcon = theme.NewThemedResource(maximizeIcon, nil)
	// IconifyIcon is the material design icon for minimizing a window
	IconifyIcon = theme.NewThemedResource(iconifyIcon, nil)
	// KeyboardIcon is the material design icon for the keyboard settings
	KeyboardIcon = theme.NewThemedResource(keyboardIcon, nil)
	// SoundIcon is the material design icon for sound in light and dark theme
	SoundIcon = theme.NewThemedResource(soundIcon, nil)
	// MuteIcon is the material design icon for mute in light and dark theme
	MuteIcon = theme.NewThemedResource(muteIcon, nil)

	// BorderWidth is the width of window frames
	BorderWidth = 4
	// ButtonWidth is the width of window buttons
	ButtonWidth = 28
	// TitleHeight is the height of a frame titleBar
	TitleHeight = 28

	// WidgetPanelBackgroundDark is the semi-transparent background matching the dark theme
	WidgetPanelBackgroundDark = color.RGBA{0x42, 0x42, 0x42, 0xcc}
	// WidgetPanelBackgroundLight is the semi-transparent background matching the light theme
	WidgetPanelBackgroundLight = color.RGBA{0xaa, 0xaa, 0xaa, 0xaa}
)

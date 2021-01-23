package theme // import "fyne.io/fynedesk/theme"

//go:generate fyne bundle -package theme -o bundled.go assets
//lint:file-ignore SA1019 These deprecation can not be fixed until fyne 2.0

import (
	"image/color"

	"fyne.io/fyne/v2/theme"
)

var (
	// PointerDefault is the standard pointer resource
	PointerDefault = resourcePointerPng

	// Background is the default background image
	Background = resourceLochfyneJpg
	// FyneAboutBackground is the image used as a background to the about screen
	FyneAboutBackground = resourceFyneaboutbgPng

	// BatteryIcon is the material design icon for battery in light and dark theme
	BatteryIcon = theme.NewThemedResource(resourceBatterySvg)
	// BrightnessIcon is the material design icon for brightness in light and dark theme
	BrightnessIcon = theme.NewThemedResource(resourceBrightnessSvg)
	// CalculateIcon is the material design icon for a calculator in light and dark theme
	CalculateIcon = theme.NewThemedResource(resourceCalculateSvg)
	// DisplayIcon is the material design icon for computer displays in light and dark theme
	DisplayIcon = theme.NewThemedResource(resourceDisplaySvg)
	// InternetIcon is the material design icon for the internet in light and dark theme
	InternetIcon = theme.NewThemedResource(resourceInternetSvg)
	// UserIcon is the material design icon for a user in light and dark theme
	UserIcon = theme.NewThemedResource(resourcePersonSvg)

	// BrokenImageIcon is the material design icon for a broken image
	BrokenImageIcon = theme.NewThemedResource(resourceBrokenimageSvg)
	// MaximizeIcon is the material design icon for maximizing a window
	MaximizeIcon = theme.NewThemedResource(resourceMaximizeSvg)
	// IconifyIcon is the material design icon for minimizing a window
	IconifyIcon = theme.NewThemedResource(resourceMinimizeSvg)
	// KeyboardIcon is the material design icon for the keyboard settings
	KeyboardIcon = theme.NewThemedResource(resourceKeyboardSvg)
	// SoundIcon is the material design icon for sound in light and dark theme
	SoundIcon = theme.NewThemedResource(resourceSoundSvg)
	// MuteIcon is the material design icon for mute in light and dark theme
	MuteIcon = theme.NewThemedResource(resourceMuteSvg)

	// BorderWidth is the width of window frames
	BorderWidth = float32(4)
	// ButtonWidth is the width of window buttons
	ButtonWidth = float32(28)
	// TitleHeight is the height of a frame titleBar
	TitleHeight = float32(28)

	// WidgetPanelBackgroundDark is the semi-transparent background matching the dark theme
	WidgetPanelBackgroundDark = color.RGBA{0x42, 0x42, 0x42, 0xcc}
	// WidgetPanelBackgroundLight is the semi-transparent background matching the light theme
	WidgetPanelBackgroundLight = color.RGBA{0xaa, 0xaa, 0xaa, 0xaa}
)

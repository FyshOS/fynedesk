package theme // import "fyshos.com/fynedesk/theme"

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ColorNamePanelBackground is used in themes to look up the background color
const ColorNamePanelBackground fyne.ThemeColorName = "fynedeskPanelBackground"

var (
	// PointerDefault is the standard pointer resource
	PointerDefault = resourcePointerPng

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
	// EthernetIcon is the material design icon for a network connection
	EthernetIcon = theme.NewThemedResource(resourceEthernetSvg)
	// WifiIcon is the material design icon for a wireless network connection
	WifiIcon = theme.NewThemedResource(resourceWifiSvg)
	// WifiOffIcon is the material design icon for a wireless device without a connection
	WifiOffIcon = theme.NewThemedResource(resourceWifioffSvg)
	// PowerIcon is the material design icon for a power connection in light and dark theme
	PowerIcon = theme.NewThemedResource(resourcePowerSvg)
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
	// LockIcon is the material design icon for the screen lock icon
	LockIcon = theme.NewThemedResource(resourceLockSvg)
	// SoundIcon is the material design icon for sound in light and dark theme
	SoundIcon = theme.NewThemedResource(resourceSoundSvg)
	// MuteIcon is the material design icon for mute in light and dark theme
	MuteIcon = theme.NewThemedResource(resourceMuteSvg)

	// BorderWidth is the width of window frames
	BorderWidth = float32(4)
	// ButtonWidth is the width of window buttons
	ButtonWidth = float32(32)
	// NarrowBarWidth is the size for the bars in narrow layout
	NarrowBarWidth = float32(36)
	// TitleHeight is the height of a frame titleBar
	TitleHeight = float32(28)
	// WidgetPanelWidth defines how wide the large widget panel should be
	WidgetPanelWidth = float32(196)
)

// WidgetPanelBackground returns the semi-transparent background matching the users current theme theme
func WidgetPanelBackground() color.Color {
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	if th := fyne.CurrentApp().Settings().Theme(); th != nil {
		col := th.Color(ColorNamePanelBackground, variant)
		if col != color.Transparent {
			return col
		}
	}

	if variant == theme.VariantLight {
		return color.RGBA{0xaa, 0xaa, 0xaa, 0xaa}
	}
	return color.RGBA{0x24, 0x24, 0x24, 0xcc}
}

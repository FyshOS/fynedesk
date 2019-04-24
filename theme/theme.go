package theme

var (
	// PointerDefault is the standard pointer resource
	PointerDefault = pointerDefault

	// BatteryIcon is the material design icon for battery in light and dark theme
	BatteryIcon = theme.NewThemedResource(batteryDark, batteryLight)

	// BorderWidth is the width of window frames
	BorderWidth = 4
	// ButtonWidth is the width of window buttons
	ButtonWidth = 24
	// TitleHeight is the height of a frame titleBar
	TitleHeight = 16
)

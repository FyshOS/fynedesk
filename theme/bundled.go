package theme

import (
	_ "embed" // embedding static files from assets directory

	"fyne.io/fyne/v2"
)

//go:embed assets/battery.svg
var batterySvg []byte

var resourceBatterySvg = &fyne.StaticResource{
	StaticName:    "battery.svg",
	StaticContent: batterySvg,
}

//go:embed assets/brightness.svg
var brightnessSvg []byte

var resourceBrightnessSvg = &fyne.StaticResource{
	StaticName:    "brightness.svg",
	StaticContent: brightnessSvg,
}

//go:embed assets/broken_image.svg
var brokenimageSvg []byte

var resourceBrokenimageSvg = &fyne.StaticResource{
	StaticName:    "broken_image.svg",
	StaticContent: brokenimageSvg,
}

//go:embed assets/calculate.svg
var calculateSvg []byte

var resourceCalculateSvg = &fyne.StaticResource{
	StaticName:    "calculate.svg",
	StaticContent: calculateSvg,
}

//go:embed assets/display.svg
var displaySvg []byte

var resourceDisplaySvg = &fyne.StaticResource{
	StaticName:    "display.svg",
	StaticContent: displaySvg,
}

//go:embed assets/ethernet.svg
var ethernetSvg []byte

var resourceEthernetSvg = &fyne.StaticResource{
	StaticName:    "ethernet.svg",
	StaticContent: ethernetSvg,
}

//go:embed assets/fyne.png
var fynePng []byte

var resourceFynePng = &fyne.StaticResource{
	StaticName:    "fyne.png",
	StaticContent: fynePng,
}

//go:embed assets/fish_onwhite.png
var fyshOSPng []byte

var resourceFyshOSPng = &fyne.StaticResource{
	StaticName:    "fish_onwhite.png",
	StaticContent: fyshOSPng,
}

//go:embed assets/icon.png
var fyneResourceIconPng []byte

var resourceIconPng = &fyne.StaticResource{
	StaticName:    "icon.png",
	StaticContent: fyneResourceIconPng,
}

//go:embed assets/internet.svg
var internetSvg []byte

var resourceInternetSvg = &fyne.StaticResource{
	StaticName:    "internet.svg",
	StaticContent: internetSvg,
}

//go:embed assets/keyboard.svg
var keyboardSvg []byte

var resourceKeyboardSvg = &fyne.StaticResource{
	StaticName:    "keyboard.svg",
	StaticContent: keyboardSvg,
}

//go:embed assets/lock.svg
var lockSvg []byte

var resourceLockSvg = &fyne.StaticResource{
	StaticName:    "lock.svg",
	StaticContent: lockSvg,
}

//go:embed assets/maximize.svg
var maximizeSvg []byte

var resourceMaximizeSvg = &fyne.StaticResource{
	StaticName:    "maximize.svg",
	StaticContent: maximizeSvg,
}

//go:embed assets/minimize.svg
var minimizeSvg []byte

var resourceMinimizeSvg = &fyne.StaticResource{
	StaticName:    "minimize.svg",
	StaticContent: minimizeSvg,
}

//go:embed assets/mute.svg
var muteSvg []byte

var resourceMuteSvg = &fyne.StaticResource{
	StaticName:    "mute.svg",
	StaticContent: muteSvg,
}

//go:embed assets/person.svg
var personSvg []byte

var resourcePersonSvg = &fyne.StaticResource{
	StaticName:    "person.svg",
	StaticContent: personSvg,
}

//go:embed assets/pointer.png
var pointerPng []byte

var resourcePointerPng = &fyne.StaticResource{
	StaticName:    "pointer.png",
	StaticContent: pointerPng,
}

//go:embed assets/power.svg
var powerSvg []byte

var resourcePowerSvg = &fyne.StaticResource{
	StaticName:    "power.svg",
	StaticContent: powerSvg,
}

//go:embed assets/sound.svg
var soundSvg []byte

var resourceSoundSvg = &fyne.StaticResource{
	StaticName:    "sound.svg",
	StaticContent: soundSvg,
}

//go:embed assets/wifi.svg
var wifiSvg []byte

var resourceWifiSvg = &fyne.StaticResource{
	StaticName:    "wifi.svg",
	StaticContent: wifiSvg,
}

//go:embed assets/wifi_off.svg
var wifiOffSvg []byte

var resourceWifioffSvg = &fyne.StaticResource{
	StaticName:    "wifi_off.svg",
	StaticContent: wifiOffSvg,
}

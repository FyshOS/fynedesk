//go:build !openbsd && !freebsd && !netbsd
// +build !openbsd,!freebsd,!netbsd

package status

import (
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
)

func (b *battery) powered() (bool, error) {
	status, err := os.ReadFile("/sys/class/power_supply/BAT0/status")
	if err != nil {
		return true, err // assume power if no battery info
	}

	return strings.ToLower(strings.TrimSpace(string(status))) != "discharging", nil
}

func (b *battery) value() (float64, error) {
	nowFile, fullFile := pickChargeOrEnergy()
	fullStr, err1 := os.ReadFile(fullFile)
	if os.IsNotExist(err1) {
		return 0, err1 // return quietly if the file was not present (desktop?)
	}
	nowStr, err2 := os.ReadFile(nowFile)
	if err1 != nil || err2 != nil {
		fyne.LogError("Error reading battery info", err1)
		return 0, err1
	}

	now, err1 := strconv.Atoi(strings.TrimSpace(string(nowStr)))
	full, err2 := strconv.Atoi(strings.TrimSpace(string(fullStr)))
	if err1 != nil || err2 != nil {
		fyne.LogError("Error converting battery info", err1)
		return 0, err1
	}

	return float64(now) / float64(full), nil
}

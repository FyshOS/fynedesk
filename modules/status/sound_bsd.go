// +build openbsd freebsd netbsd

package status

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	wmtheme "fyne.io/fynedesk/theme"
)

var volume int

// Destroy destroys the module
func (b *sound) Destroy() {
	if b.client == nil {
		return
	}
	b.client.Close()
}

func (b *sound) value() (int, error) {
	cmd := exec.Command("mixer", "vol")
	out, err := cmd.Output()
	colonPos := strings.Index(out, ":")
	if err != nil || colonPos <= 0 {
		log.Println("Failed to get volume", err)
		return
	}

	volume, err := strconv.atoi(out[:colonPos])
	return volume, err
}

func (b *sound) offsetValue(diff int) {
	floatVal, err := b.value()
	if err != nil {
		log.Println("Failed to get volume", err)
		return
	}
	value := floatVal + float32(diff)/100

	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}

	b.setValue(value)
}

func (b *sound) setup() error {
	_, err := b.value()
	return err
}

func (b *sound) setValue(vol int) {
	level := fmt.Sprintf("%d:%d", vol, vol)
	cmd := exec.Command("mixer", "vol", level)
	if err := cmd.Run(); err != nil {
		log.Println("Failed to set volume", err)
		return
	}

	volume = vol
	b.bar.SetValue(float64(vol))
}

func (b *sound) toggleMute() {
	oldVal, err := b.value()
	if err != nil {
		log.Println("toggleMute() failed", err)
		return
	}

	mute := oldVal != 0
	if mute {
		b.setValue(0)
	} else {
		b.setValue(volume)
	}

	if mute {
		b.mute.SetIcon(wmtheme.MuteIcon)
	} else {
		b.mute.SetIcon(wmtheme.SoundIcon)
	}
}

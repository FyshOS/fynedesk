// +build openbsd freebsd netbsd

package status

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"strconv"

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
	cmd := exec.Command("mixer", "-s", "vol")
	out, err := cmd.Output()
	colonPos := strings.Index(string(out), ":")
	if err != nil || colonPos <= 0 {
		return 0, err
	}

	// strip "vol " from the start as well
	volume, err := strconv.atoi(string(out[4:colonPos]))
	return volume, err
}

func (b *sound) setup() error {
	_, err := b.value()
	if volume == 0 {
		volume = 75 // we loaded in mute mode perhapse
	}
	return err
}

func (b *sound) setValue(vol int) {
	level := fmt.Sprintf("%d:%d", vol, vol)
	cmd := exec.Command("mixer", "vol", level)
	if err := cmd.Run(); err != nil {
		log.Println("Failed to set volume", err)
		return
	}

	b.bar.SetValue(float64(vol))
	if vol == 0 {
		b.mute.SetIcon(wmtheme.MuteIcon)
	} else {
		volume = vol
		b.mute.SetIcon(wmtheme.SoundIcon)
	}
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
}

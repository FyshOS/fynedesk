//go:build openbsd || freebsd || netbsd
// +build openbsd freebsd netbsd

package status

import (
	"bytes"
	"os/exec"
	"strconv"

	"fyne.io/fyne/v2"
)

var (
	muted  bool
	volume int
)

// Destroy tidies up resources
func (b *sound) Destroy() {
	if b.client == nil {
		return
	}
	b.client.Close()
}

func (b *sound) muted() bool {
	return muted
}

func (b *sound) value() (int, error) {
	cmd := exec.Command("mixer", "-s", "vol")
	out, err := cmd.Output()
	colonPos := bytes.IndexByte(out, ':')
	if err != nil || colonPos <= 0 {
		return 0, err
	}

	// strip "vol " from the start as well
	volume, err := strconv.Atoi(string(out[4:colonPos]))
	return volume, err
}

func (b *sound) setup() error {
	_, err := b.value()
	if volume == 0 {
		muted = true
		volume = 75 // we loaded in mute mode perhapse
	}
	return err
}

func (b *sound) setValue(vol int) {
	volStr := strconv.Itoa(vol)
	level := volStr + ":" + volStr
	cmd := exec.Command("mixer", "vol", level)
	if err := cmd.Run(); err != nil {
		fyne.LogError("Failed to set volume", err)
		return
	}

	if b.bar != nil { // no status area: the mixer is still set, there is just nothing to draw
		b.bar.SetValue(float64(vol))
	}
	b.updateIcon(vol, muted)
}

func (b *sound) setMute(mute bool) {
	if mute == muted {
		return
	}
	b.toggleMute()
}

func (b *sound) toggleMute() {
	muted = !muted
	if muted {
		b.setValue(0)
	} else {
		b.setValue(volume)
	}
}

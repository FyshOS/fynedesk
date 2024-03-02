//go:build !openbsd && !freebsd && !netbsd
// +build !openbsd,!freebsd,!netbsd

package status

import (
	"fyne.io/fyne/v2"

	"github.com/mafik/pulseaudio"

	wmtheme "fyshos.com/fynedesk/theme"
)

// Destroy tidies up resources
func (b *sound) Destroy() {
	if b.client == nil {
		return
	}
	b.client.Close()
}

func (b *sound) muted() bool {
	m, _ := b.client.Mute()
	return m
}

func (b *sound) value() (int, error) {
	volume, err := b.client.Volume()
	if err != nil {
		return 0, err
	}

	return int(volume * 100), nil
}

func (b *sound) setup() error {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return err
	}
	b.client = client
	return nil
}

func (b *sound) setValue(vol int) {
	if err := b.client.SetVolume(float32(vol) / 100); err != nil {
		fyne.LogError("Failed to set volume", err)
		return
	}

	b.bar.SetValue(float64(vol))
}

func (b *sound) toggleMute() {
	toggle, err := b.client.ToggleMute()
	if err != nil {
		fyne.LogError("toggleMute() failed", err)
		return
	}

	if toggle {
		b.mute.SetIcon(wmtheme.MuteIcon)
	} else {
		b.mute.SetIcon(wmtheme.SoundIcon)
	}

}

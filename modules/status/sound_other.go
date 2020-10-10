// +build !openbsd,!freebsd,!netbsd

package status

import (
	"log"

	wmtheme "fyne.io/fynedesk/theme"
	"github.com/mafik/pulseaudio"
)

// Destroy destroys the module
func (b *sound) Destroy() {
	if b.client == nil {
		return
	}
	b.client.Close()
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
		log.Println("Failed to set volume", err)
		return
	}

	b.bar.SetValue(float64(vol))
}

func (b *sound) toggleMute() {
	toggl, err := b.client.ToggleMute()
	if err != nil {
		log.Println("toggleMute() failed", err)
		return
	}

	if toggl {
		b.mute.SetIcon(wmtheme.MuteIcon)
	} else {
		b.mute.SetIcon(wmtheme.SoundIcon)
	}

}

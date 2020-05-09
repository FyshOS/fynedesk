package builtin

import (
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/mattboll/pulseaudio"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

var pulseaudioMeta = fynedesk.ModuleMetadata{
	Name:        "Pulseaudio",
	NewInstance: NewPulseaudio,
}

// Pulseaudio is a progress bar module to modify screen pulseaudio
type Pulseaudio struct {
	bar       *widget.ProgressBar
	client    *pulseaudio.Client
	soundIcon *widget.Button
}

// Destroy destroys the module
func (b *Pulseaudio) Destroy() {
	b.client.Close()
}

func (b *Pulseaudio) value() (float32, bool, error) {
	volume, err := b.client.Volume()
	if err != nil {
		return 0, true, err
	}

	return volume, false, nil
}

// OffsetValue actually increase or decrease the screen pulseaudio
func (b *Pulseaudio) OffsetValue(diff int) {
	floatVal, _, _ := b.value()
	value := floatVal + float32(diff)/100

	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}

	if err := b.client.SetVolume(value); err != nil {
		log.Println("Failed to set volume", err)
		return
	}

	b.bar.SetValue(float64(value))
}

// ToggleMute is used to switch between mute or not
func (b *Pulseaudio) ToggleMute() {
	toggl, err := b.client.ToggleMute()
	if err != nil {
		log.Println("ToggleMute() failed", err)
		return
	}

	if toggl {
		b.soundIcon.SetIcon(wmtheme.MuteIcon)
	} else {
		b.soundIcon.SetIcon(wmtheme.SoundIcon)
	}

}

// StatusAreaWidget builds the widget
func (b *Pulseaudio) StatusAreaWidget() fyne.CanvasObject {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return nil
	}
	b.client = client

	b.bar = widget.NewProgressBar()
	b.soundIcon = widget.NewButtonWithIcon("", wmtheme.SoundIcon, func() {
		b.ToggleMute()
	})
	less := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		b.OffsetValue(-5)
	})
	more := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		b.OffsetValue(5)
	})
	sound := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, less, more),
		less, b.bar, more)

	go b.OffsetValue(0)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, b.soundIcon, nil), b.soundIcon, sound)
}

// Metadata returns ModuleMetadata
func (b *Pulseaudio) Metadata() fynedesk.ModuleMetadata {
	return pulseaudioMeta
}

// NewPulseaudio creates a new module that will show screen pulseaudio in the status area
func NewPulseaudio() fynedesk.Module {
	return &Pulseaudio{}
}

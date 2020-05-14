package builtin

import (
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/mafik/pulseaudio"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

var soundMeta = fynedesk.ModuleMetadata{
	Name:        "Sound",
	NewInstance: newSound,
}

type sound struct {
	bar    *widget.ProgressBar
	client *pulseaudio.Client
	mute   *widget.Button
}

// Destroy destroys the module
func (b *sound) Destroy() {
	b.client.Close()
}

func (b *sound) value() (float32, error) {
	volume, err := b.client.Volume()
	if err != nil {
		return 0, err
	}

	return volume, nil
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

	if err := b.client.SetVolume(value); err != nil {
		log.Println("Failed to set volume", err)
		return
	}

	b.bar.SetValue(float64(value))
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

// StatusAreaWidget builds the widget
func (b *sound) StatusAreaWidget() fyne.CanvasObject {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return nil
	}
	b.client = client

	b.bar = widget.NewProgressBar()
	b.mute = widget.NewButtonWithIcon("", wmtheme.SoundIcon, func() {
		b.toggleMute()
	})
	less := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		b.offsetValue(-5)
	})
	more := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		b.offsetValue(5)
	})
	sound := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, less, more),
		less, b.bar, more)

	go b.offsetValue(0)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, b.mute, nil), b.mute, sound)
}

// Metadata returns ModuleMetadata
func (b *sound) Metadata() fynedesk.ModuleMetadata {
	return soundMeta
}

func newSound() fynedesk.Module {
	return &sound{}
}

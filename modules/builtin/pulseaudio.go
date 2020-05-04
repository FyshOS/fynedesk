package builtin

import (
	"errors"
	"log"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/auroralaboratories/pulse"

	"fyne.io/fynedesk"
	wmtheme "fyne.io/fynedesk/theme"
)

var pulseaudioMeta = fynedesk.ModuleMetadata{
	Name:        "Pulseaudio",
	NewInstance: NewPulseaudio,
}

// Pulseaudio is a progress bar module to modify screen pulseaudio
type Pulseaudio struct {
	bar    *widget.ProgressBar
	client *pulse.Client
}

// Destroy destroys the module
func (b *Pulseaudio) Destroy() {
	b.client.Destroy()
}

func (b *Pulseaudio) value() (float64, bool, error) {
	sinks, err := b.client.GetSinks()
	if err != nil {
		return 0, true, err
	}
	if len(sinks) <= 0 {
		return 0, true, errors.New("no sinks returned")
	}

	sink := sinks[0]
	if err := sink.Refresh(); err != nil {
		return 0, true, err
	}

	return sink.VolumeFactor, sink.Muted, nil
}

// OffsetValue actually increase or decrease the screen pulseaudio
func (b *Pulseaudio) OffsetValue(diff int) {
	sinks, err := b.client.GetSinks()
	if err != nil || len(sinks) <= 0 {
		log.Println("GetSinks() failed", err)
		return
	}

	sink := sinks[0]

	floatVal, _, _ := b.value()
	value := floatVal + float64(diff)/100

	if value < 0 {
		value = 0
	} else if value > 1 {
		value = 1
	}

	if err := sink.SetVolume(value); err != nil {
		log.Println("Failed to set volume", err)
		return
	}

	b.bar.SetValue(value)
}

// StatusAreaWidget builds the widget
func (b *Pulseaudio) StatusAreaWidget() fyne.CanvasObject {
	client, err := pulse.NewClient(`fyne-client`)
	if err != nil {
		return nil
	}
	b.client = client

	b.bar = widget.NewProgressBar()
	soundIcon := widget.NewIcon(wmtheme.SoundIcon)
	less := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		b.OffsetValue(-5)
	})
	more := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		b.OffsetValue(5)
	})
	sound := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, less, more),
		less, b.bar, more)

	go b.OffsetValue(0)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, soundIcon, nil), soundIcon, sound)
}

// Metadata returns ModuleMetadata
func (b *Pulseaudio) Metadata() fynedesk.ModuleMetadata {
	return pulseaudioMeta
}

// NewPulseaudio creates a new module that will show screen pulseaudio in the status area
func NewPulseaudio() fynedesk.Module {
	return &Pulseaudio{}
}

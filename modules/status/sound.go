package status

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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

func newSound() fynedesk.Module {
	return &sound{}
}

// StatusAreaWidget builds the widget
func (b *sound) StatusAreaWidget() fyne.CanvasObject {
	if err := b.setup(); err != nil {
		fyne.LogError("Unable to start sound module", err)
		return nil
	}

	b.bar = &widget.ProgressBar{Max: 100}
	b.mute = &widget.Button{Icon: wmtheme.SoundIcon, Importance: widget.LowImportance, OnTapped: b.toggleMute}

	less := &widget.Button{Icon: theme.ContentRemoveIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(-5)
	}}

	more := &widget.Button{Icon: theme.ContentAddIcon(), Importance: widget.LowImportance, OnTapped: func() {
		b.offsetValue(5)
	}}

	sound := container.NewBorder(nil, nil, less, more, b.bar)

	go b.offsetValue(0)
	return container.NewBorder(nil, nil, b.mute, nil, sound)
}

// Metadata returns ModuleMetadata
func (b *sound) Metadata() fynedesk.ModuleMetadata {
	return soundMeta
}

func (b *sound) offsetValue(diff int) {
	currVal, err := b.value()
	if err != nil {
		fyne.LogError("Failed to get volume", err)
		return
	}
	value := currVal + diff

	if value < 0 {
		value = 0
	} else if value > 100 {
		value = 100
	}

	b.setValue(value)
}

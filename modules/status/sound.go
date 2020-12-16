package status

import (
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

	sound := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, less, more), less, b.bar, more)

	go b.offsetValue(0)
	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, b.mute, nil), b.mute, sound)
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

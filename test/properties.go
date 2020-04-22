package test

import (
	"fyne.io/fyne"
)

type dummyProperties struct {
	name, cmd, iconName string
	class               []string
}

func (d dummyProperties) Class() []string {
	return d.class
}

func (d dummyProperties) Command() string {
	return d.cmd
}

func (d dummyProperties) Decorated() bool {
	return true
}

func (d dummyProperties) Icon() fyne.Resource {
	return nil
}

func (d dummyProperties) IconName() string {
	return d.iconName
}

func (d dummyProperties) SkipTaskbar() bool {
	return false
}

func (d dummyProperties) Title() string {
	return d.name
}

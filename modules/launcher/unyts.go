package launcher

import (
	_ "embed" // embed icon
	"log"
	"os/exec"
	"strconv"
	"strings"

	"codeberg.org/sdassow/unyts/units"
	"fyne.io/fyne/v2"
	"fyshos.com/fynedesk"
)

var unytsMeta = fynedesk.ModuleMetadata{
	Name:        "Launcher: Convert units",
	NewInstance: newUnyts,
}

//go:embed unyts.png
var resourceUnytsPngData []byte

var resourceUnyts = &fyne.StaticResource{
	StaticName:    "unyts.png",
	StaticContent: resourceUnytsPngData,
}

type unyts struct {
}

func (u *unyts) Destroy() {
}

func (u *unyts) LaunchSuggestions(input string) []fynedesk.LaunchSuggestion {
	if u.isConversion(input) {
		return []fynedesk.LaunchSuggestion{&unytResult{input}}
	}

	return nil
}

func (u *unyts) Metadata() fynedesk.ModuleMetadata {
	return urlMeta
}

func (u *unyts) isConversion(input string) bool {
	parts := strings.Split(input, " ")
	log.Println("count", len(parts))
	if len(parts) != 3 {
		return false
	}

	pivot := strings.ToLower(parts[1])
	return pivot == "in" || pivot == "as" || pivot == "to"
}

// newURLs creates a new module that will show URLs in the launcher suggestions
func newUnyts() fynedesk.Module {
	return &unyts{}
}

type unytResult struct {
	conversion string
}

func (r *unytResult) Icon() fyne.Resource {
	return resourceUnyts
}

func (r *unytResult) Title() string {
	parts := strings.Split(r.conversion, " ")
	src, unit, err := units.ParseAmountAndUnit(parts[0])
	if err != nil {
		return err.Error()
	}

	lookup := units.FuncMapForAbbr(unit.Abbr)
	conv, ok := lookup[parts[2]]
	if !ok {
		return "Cannot convert " + unit.Abbr + " to " + parts[2]
	}

	return r.conversion + " = " + strconv.FormatFloat(conv(src), 'f', 1, 64)
}

func (r *unytResult) Launch() {
	_ = exec.Command("unyts", "-c", r.conversion).Run()
}

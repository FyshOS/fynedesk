package launcher

import (
	"regexp"
	"strconv"

	"fyne.io/fyne"

	"fyne.io/fynedesk"
	wmTheme "fyne.io/fynedesk/theme"

	"github.com/Knetic/govaluate"
)

var calcMeta = fynedesk.ModuleMetadata{
	Name:        "Launch Calculations",
	NewInstance: newCalcSuggest,
}

type calc struct {
	exprRegex, numRegex *regexp.Regexp
}

func (c *calc) Destroy() {
}

func (c *calc) LaunchSuggestions(input string) []fynedesk.LaunchSuggestion {
	if !c.isExpression(input) {
		return nil
	}

	result, err := c.eval(input)
	if err != nil {
		return nil
	}
	return []fynedesk.LaunchSuggestion{&calcItem{sum: input, result: result}}
}

func (c *calc) eval(sum string) (string, error) {
	expression, err := govaluate.NewEvaluableExpression(sum)
	if err != nil {
		return "", err
	}

	res, err := expression.Evaluate(nil)
	if err != nil {
		return "", err
	}

	if f, ok := res.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	}

	return "", nil
}

func (c *calc) Metadata() fynedesk.ModuleMetadata {
	return calcMeta
}

// isExpression will return true if input is a mathematical expression unless it just contains a number
func (c *calc) isExpression(input string) bool {
	return c.exprRegex.MatchString(input) && !c.numRegex.MatchString(input)
}

// newURLs creates a new module that will show URLs in the launcher suggestions
func newCalcSuggest() fynedesk.Module {
	expr, _ := regexp.Compile("^[0-9.+\\-*/()]+$")
	num, _ := regexp.Compile("^[0-9.]+$")
	return &calc{exprRegex: expr, numRegex: num}
}

type calcItem struct {
	sum, result string
}

func (i *calcItem) Icon() fyne.Resource {
	return wmTheme.CalculateIcon
}

func (i *calcItem) Title() string {
	return i.sum + " = " + i.result
}

func (i *calcItem) Launch() {
	fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(i.result)
}

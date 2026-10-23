package ai

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyshos.com/tyde/internal/modhelper"

	"fyshos.com/tyde"
	wmTest "fyshos.com/tyde/test"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// commandModule is a module that both describes a command and answers it, as a
// real one does.
type commandModule struct {
	launched string
}

func (m *commandModule) Destroy() {}

func (m *commandModule) Metadata() tyde.ModuleMetadata {
	return tyde.ModuleMetadata{Name: "Test Commands"}
}

func (m *commandModule) Commands() []tyde.Command {
	c := tyde.Command{
		Name:        "brightness",
		Description: "Set the screen brightness",
		Args:        "a percentage",
		Examples:    []string{"brightness 50"},
	}

	modhelper.AnswerWith(&c, m.item)
	return []tyde.Command{c}
}

func (m *commandModule) LaunchSuggestions(input string) []tyde.LaunchSuggestion {
	return []tyde.LaunchSuggestion{&commandItem{m: m, input: input}}
}

// item is the seam a real module shares between its launcher entry and its
// command - the same item answers both.
func (m *commandModule) item(args string) tyde.LaunchSuggestion {
	if args == "" {
		return nil
	}
	return &commandItem{m: m, input: "brightness " + args}
}

type commandItem struct {
	m     *commandModule
	input string
}

func (i *commandItem) Icon() fyne.Resource { return nil }
func (i *commandItem) Title() string       { return "ran " + i.input }
func (i *commandItem) Launch()             { i.m.launched = i.input }

// withCommandModule points the desktop at a module offering one command.
func withCommandModule(t *testing.T) *commandModule {
	t.Helper()
	test.NewApp()

	mod := &commandModule{}
	desk := wmTest.NewDesktop()
	desk.SetModules([]tyde.Module{mod})
	tyde.SetInstance(desk)
	t.Cleanup(func() { tyde.SetInstance(nil) })
	return mod
}

func TestDesktopTools(t *testing.T) {
	withCommandModule(t)

	tools := desktopTools()
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Function.Name)
	}
	assert.Contains(t, names, "tyde_brightness") // the module's command
	assert.Contains(t, names, "tyde_welcome")    // a desktop built-in

	var brightness *llms.FunctionDefinition
	for _, tool := range tools {
		if tool.Function.Name == "tyde_brightness" {
			brightness = tool.Function
		}
	}
	if assert.NotNil(t, brightness) {
		assert.Equal(t, "Set the screen brightness. Examples: brightness 50", brightness.Description)
		params := brightness.Parameters.(map[string]any)
		props := params["properties"].(map[string]any)
		assert.Equal(t, "a percentage", props[argsField].(map[string]any)["description"])
	}
}

// A command taking no arguments describes none, so the model does not invent
// any to pass.
func TestDesktopTools_NoArgs(t *testing.T) {
	withCommandModule(t)

	for _, tool := range desktopTools() {
		if tool.Function.Name != "tyde_mute" && tool.Function.Name != "tyde_restart" {
			continue
		}
		params := tool.Function.Parameters.(map[string]any)
		assert.Empty(t, params["properties"])
	}
}

func TestDesktopTools_Disabled(t *testing.T) {
	withCommandModule(t)
	fyne.CurrentApp().Preferences().SetBool(prefTools, false)

	assert.Nil(t, desktopTools())
}

func TestRunTool(t *testing.T) {
	mod := withCommandModule(t)

	line, result := runTool("tyde_brightness", `{"args": "50"}`)
	assert.Equal(t, "brightness 50", line)
	assert.Equal(t, "ran brightness 50", result)
	assert.Equal(t, "brightness 50", mod.launched)
}

func TestRunTool_UnknownCommand(t *testing.T) {
	withCommandModule(t)

	line, result := runTool("tyde_selfdestruct", "{}")
	assert.Empty(t, line)
	assert.Equal(t, `error: no such command "selfdestruct"`, result)
}

func TestRunTool_Disabled(t *testing.T) {
	mod := withCommandModule(t)
	fyne.CurrentApp().Preferences().SetBool(prefTools, false)

	_, result := runTool("tyde_brightness", `{"args": "50"}`)
	assert.Contains(t, result, "not allowed")
	assert.Empty(t, mod.launched)
}

func TestToolName(t *testing.T) {
	assert.Equal(t, "tyde_brightness", toolName("brightness"))
	assert.Equal(t, "tyde_large_type", toolName("large type"))
	assert.Equal(t, "tyde_2_2", toolName("2+2"))
}

func TestToolArgs(t *testing.T) {
	assert.Equal(t, "50", toolArgs(`{"args": "50"}`))
	assert.Equal(t, "50", toolArgs(`{"args": 50}`)) // sent as a number
	assert.Equal(t, "0.5", toolArgs(`{"args": 0.5}`))
	assert.Empty(t, toolArgs(""))
	assert.Empty(t, toolArgs("{}"))
	assert.Empty(t, toolArgs("not json"))
	assert.Empty(t, toolArgs(`{"other": "50"}`))
}

func TestFirstToolCall(t *testing.T) {
	assert.Nil(t, firstToolCall(nil))
	assert.Nil(t, firstToolCall(&llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: "just words"}},
	}))

	// Anthropic splits words and a call across choices - the call is in the second.
	call := firstToolCall(&llms.ContentResponse{Choices: []*llms.ContentChoice{
		{Content: "I'll do that"},
		{ToolCalls: []llms.ToolCall{{
			ID:           "call-1",
			FunctionCall: &llms.FunctionCall{Name: "tyde_brightness", Arguments: `{"args":"50"}`},
		}}},
	}})
	if assert.NotNil(t, call) {
		assert.Equal(t, "call-1", call.ID)
		assert.Equal(t, "tyde_brightness", call.FunctionCall.Name)
	}
}

func TestLooksLikeToolCall(t *testing.T) {
	assert.True(t, looksLikeToolCall(`[{"id":"call_1","type":"function"`)) // still streaming
	assert.True(t, looksLikeToolCall("\n {\"name\""))
	assert.False(t, looksLikeToolCall("Setting the brightness to 50%"))
	assert.False(t, looksLikeToolCall("```json\n{"))
}

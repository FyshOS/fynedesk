package ai

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"

	"fyshos.com/tyde"
	"fyshos.com/tyde/modules/rpc"

	"github.com/tmc/langchaingo/llms"
)

const (
	toolPrefix   = "tyde_"
	argsField    = "args"
	maxToolTurns = 4
	chatCommand  = "fathom"
)

// toolsEnabled reports whether the assistant may run desktop commands.
func toolsEnabled() bool {
	return fyne.CurrentApp().Preferences().BoolWithFallback(prefTools, true)
}

// desktopTools is commandTools unless the person has turned tool use off.
func desktopTools() []llms.Tool {
	if !toolsEnabled() {
		return nil
	}
	return commandTools()
}

// commandTools describes the commands the desktop currently offers as model tools.
func commandTools() []llms.Tool {
	var tools []llms.Tool
	seen := map[string]bool{}
	for _, cmd := range rpc.Commands() {
		if cmd.Name == chatCommand {
			continue // opening the assistant is no use to the assistant
		}

		name := toolName(cmd.Name)
		if name == toolPrefix || seen[name] {
			continue // unnamed, or a duplicate a provider would reject
		}
		seen[name] = true

		properties := map[string]any{}
		if cmd.Args != "" {
			properties[argsField] = map[string]any{
				"type":        "string",
				"description": cmd.Args,
			}
		}

		tools = append(tools, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        name,
				Description: describe(cmd),
				Parameters: map[string]any{
					"type":       "object",
					"properties": properties,
				},
			},
		})
	}
	return tools
}

// describe writes the one line a model sees for a command, adding the examples
// as the clearest statement of the form the arguments take.
func describe(cmd tyde.Command) string {
	desc := cmd.Description
	if len(cmd.Examples) > 0 {
		desc += ". Examples: " + strings.Join(cmd.Examples, ", ")
	}
	return desc
}

// toolName maps a command name to a tool name, which providers limit to
// letters, digits, underscores and dashes.
func toolName(command string) string {
	return toolPrefix + strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
			return r
		default:
			return '_'
		}
	}, command)
}

// runTool performs a tool call from the model: it rebuilds the command line the
// tool stands for and runs it exactly as tyde_ctl would, returning the cmd as output.
func runTool(name, arguments string) (line, result string) {
	if !toolsEnabled() {
		return "", "error: the assistant is not allowed to run desktop commands"
	}

	for _, cmd := range rpc.Commands() {
		if toolName(cmd.Name) != name {
			continue
		}

		// The command found here is the one that runs: it carries its own action,
		// so nothing is turned back into text for a module to match a second time.
		args := toolArgs(arguments)
		line = strings.TrimSpace(cmd.Name + " " + args)
		result, err := rpc.RunCommand(cmd, args)
		if err != nil {
			return line, "error: " + err.Error()
		}
		if result == "" {
			result = "done"
		}
		return line, result
	}

	return "", fmt.Sprintf("error: no such command %q", strings.TrimPrefix(name, toolPrefix))
}

// toolArgs pulls the command line's tail out of the JSON arguments object the
// model sends. Anything unparseable is ignored rather than guessed.
func toolArgs(arguments string) string {
	if strings.TrimSpace(arguments) == "" {
		return ""
	}

	var fields map[string]any
	if err := json.Unmarshal([]byte(arguments), &fields); err != nil {
		return ""
	}
	switch v := fields[argsField].(type) {
	case string:
		return v
	case float64: // "brightness 50" arriving as a number rather than a string
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return ""
}

// looksLikeToolCall reports whether streamed text is a tool call being spelled
// out rather than something to read. An answer does not open with a brace,
// so the first character tells them apart.
func looksLikeToolCall(text string) bool {
	trimmed := strings.TrimLeft(text, " \t\r\n")
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

// usedTools reports whether a conversation has already run a command, and so
// carries a call the model's next request must still be able to match to a tool.
func usedTools(hist []llms.MessageContent) bool {
	for _, msg := range hist {
		for _, part := range msg.Parts {
			if _, ok := part.(llms.ToolCall); ok {
				return true
			}
		}
	}
	return false
}

// firstToolCall returns the first tool call a response asks for, if any.
func firstToolCall(resp *llms.ContentResponse) *llms.ToolCall {
	if resp == nil {
		return nil
	}
	// Anthropic splits a reply that both speaks and calls a tool across two
	// choices (the text in one, the call in another), so look through them all.
	for _, choice := range resp.Choices {
		for _, call := range choice.ToolCalls {
			if call.FunctionCall != nil {
				return &call
			}
		}
	}
	return nil
}

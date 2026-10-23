package ai

import (
	"context"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// scriptedLLM answers with the replies it was given, one per request, recording
// what it was asked each time. It stands in for a provider so the tool loop can
// be run without a server.
type scriptedLLM struct {
	replies []*llms.ContentResponse
	asked   [][]llms.MessageContent
	tools   [][]llms.Tool
}

func (l *scriptedLLM) Call(context.Context, string, ...llms.CallOption) (string, error) {
	return "", nil
}

func (l *scriptedLLM) GenerateContent(_ context.Context, messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	opts := llms.CallOptions{}
	for _, o := range options {
		o(&opts)
	}
	l.asked = append(l.asked, messages)
	l.tools = append(l.tools, opts.Tools)

	reply := l.replies[0]
	l.replies = l.replies[1:]
	return reply, nil
}

func toolReply(name, arguments string) *llms.ContentResponse {
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{
		// Local servers spell the call out in the content as well, which must not
		// reach the transcript.
		Content: `[{"id":"call-1","type":"function","function":{"name":"` + name + `"}}]`,
		ToolCalls: []llms.ToolCall{{
			ID:           "call-1",
			FunctionCall: &llms.FunctionCall{Name: name, Arguments: arguments},
		}},
	}}}
}

// transcript reads back the lines the chat has shown.
func transcript(c *chatUI) []string {
	var lines []string
	for _, obj := range c.log.Objects {
		lines = append(lines, strings.TrimSpace(text(obj)))
	}
	return lines
}

func text(obj fyne.CanvasObject) string {
	switch o := obj.(type) {
	case *widget.RichText:
		return o.String()
	case *fyne.Container:
		out := ""
		for _, child := range o.Objects {
			out += text(child)
		}
		return out
	}
	return ""
}

// newTestChat builds a chat view ready to converse, without a window.
func newTestChat() *chatUI {
	c := newChatUI()
	c.build()
	return c
}

// TestConverse_RunsCommand walks the whole loop: the model asks for a command,
// the desktop runs it, and the result goes back for the model to answer from.
func TestConverse_RunsCommand(t *testing.T) {
	mod := withCommandModule(t)
	chat := newTestChat()

	llm := &scriptedLLM{replies: []*llms.ContentResponse{
		toolReply("tyde_brightness", `{"args": "50"}`),
		{Choices: []*llms.ContentChoice{{Content: "Brightness is now 50%."}}},
	}}

	hist := chat.converse(context.Background(), llm,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "dim the screen to half")})

	assert.Equal(t, "brightness 50", mod.launched)
	assert.Empty(t, llm.replies) // both turns were used
	assert.NotEmpty(t, llm.tools[0])

	// The kept conversation is: the question, the command, its result, the answer.
	if assert.Len(t, hist, 4) {
		call, ok := hist[1].Parts[0].(llms.ToolCall)
		if assert.True(t, ok) {
			assert.Equal(t, "tyde_brightness", call.FunctionCall.Name)
		}
		result, ok := hist[2].Parts[0].(llms.ToolCallResponse)
		if assert.True(t, ok) {
			assert.Equal(t, llms.ChatMessageTypeTool, hist[2].Role)
			assert.Equal(t, "call-1", result.ToolCallID)
			assert.Equal(t, "ran brightness 50", result.Content)
		}
		assert.Equal(t, llms.ChatMessageTypeAI, hist[3].Role)
	}

	// The second request carried the command and its result back to the model.
	assert.Len(t, llm.asked[1], 3)

	// The transcript is the command that ran and the answer - no row for the turn
	// that asked for it, no repeat of what the command returned, and none of the
	// raw call the server also sent as text.
	assert.Equal(t, []string{
		"tyde_ctl brightness 50",
		"Brightness is now 50%.",
	}, transcript(chat))
}

// A command that fails says so on its own line, since the model is not told
// anything the reader cannot see.
func TestConverse_ShowsFailedCommand(t *testing.T) {
	withCommandModule(t)
	chat := newTestChat()

	llm := &scriptedLLM{replies: []*llms.ContentResponse{
		toolReply("tyde_nonsense", "{}"),
		{Choices: []*llms.ContentChoice{{Content: "I can't do that."}}},
	}}

	chat.converse(context.Background(), llm,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "do something odd")})

	assert.Equal(t, []string{
		`no such command "nonsense"`,
		"I can't do that.",
	}, transcript(chat))
}

// A plain answer leaves the desktop alone and keeps one assistant turn.
func TestConverse_NoCommand(t *testing.T) {
	mod := withCommandModule(t)
	chat := newTestChat()

	llm := &scriptedLLM{replies: []*llms.ContentResponse{
		{Choices: []*llms.ContentChoice{{Content: "Hello."}}},
	}}

	hist := chat.converse(context.Background(), llm,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	assert.Empty(t, mod.launched)
	if assert.Len(t, hist, 2) {
		assert.Equal(t, llms.ChatMessageTypeAI, hist[1].Role)
	}
}

// A model that only ever asks for commands is stopped rather than looped with.
func TestConverse_StopsRunawayTools(t *testing.T) {
	withCommandModule(t)
	chat := newTestChat()

	llm := &scriptedLLM{}
	for i := 0; i < maxToolTurns; i++ {
		llm.replies = append(llm.replies, toolReply("tyde_brightness", `{"args": "50"}`))
	}

	chat.converse(context.Background(), llm,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "brighter")})

	assert.Empty(t, llm.replies)
	assert.Len(t, llm.asked, maxToolTurns)
}

// Turning tool use off part-way through a conversation must not break it: the
// tools stay declared so the command already in the history still matches one,
// but the assistant will no longer act.
func TestConverse_ToolsDisabledMidConversation(t *testing.T) {
	mod := withCommandModule(t)
	chat := newTestChat()
	fyne.CurrentApp().Preferences().SetBool(prefTools, false)

	hist := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "dim the screen"),
		{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.ToolCall{
			ID:           "call-1",
			FunctionCall: &llms.FunctionCall{Name: "tyde_brightness", Arguments: `{"args":"50"}`},
		}}},
		{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{llms.ToolCallResponse{
			ToolCallID: "call-1", Name: "tyde_brightness", Content: "ran brightness 50",
		}}},
	}

	llm := &scriptedLLM{replies: []*llms.ContentResponse{
		toolReply("tyde_brightness", `{"args": "70"}`),
		{Choices: []*llms.ContentChoice{{Content: "I can't run commands now."}}},
	}}

	chat.converse(context.Background(), llm, hist)

	assert.NotEmpty(t, llm.tools[0]) // still declared, so the history is valid
	assert.Empty(t, mod.launched)    // but nothing was run
}

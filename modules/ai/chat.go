package ai

import (
	"context"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/tyde/modules/launcher"
	wmtheme "fyshos.com/tyde/theme"

	"github.com/tmc/langchaingo/llms"
)

const systemPrompt = "You are an AI assistant called Fathom, part of the FyshOS desktop. " +
	"Answer concisely and format replies in Markdown when it helps readability. " +
	"The tools you are given run this desktop's own commands: use them when asked to " +
	"change a setting or show something on screen, then say what happened in a sentence. " +
	"Talk about the commands by what they do rather than by their tool names."

// chatUI is the conversation view: a scrolling transcript above an input row.
// history holds the running conversation passed to the model on each turn.
type chatUI struct {
	win fyne.Window

	history   []llms.MessageContent
	log       *fyne.Container
	scroll    *container.Scroll
	entry     *widget.Entry
	send      *widget.Button
	reasoning *widget.Check // Local AI: quick per-tab speed/accuracy toggle
	busy      bool

	// isActive reports whether this chat's tab is the visible one, so a reply
	// completing in a background tab does not steal keyboard focus.
	isActive func() bool
	cancel   context.CancelFunc

	titleSetter func(string)
}

func newChatUI() *chatUI {
	return &chatUI{
		history: []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		},
	}
}

// build assembles the chat window content.
func (c *chatUI) build() fyne.CanvasObject {
	c.log = container.NewVBox()
	c.scroll = container.NewVScroll(c.log)

	c.entry = widget.NewMultiLineEntry()
	c.entry.SetPlaceHolder("Ask the assistant...")
	c.entry.Wrapping = fyne.TextWrapWord
	c.entry.OnSubmitted = func(string) { c.submit() }

	c.send = &widget.Button{
		Icon: theme.MailSendIcon(), Importance: widget.HighImportance,
		OnTapped: c.submit,
	}

	// A compact reasoning toggle above the input makes it easy to trade speed for
	// accuracy per question.
	c.reasoning = widget.NewCheck("Reasoning", nil)

	input := container.NewBorder(nil, nil, nil, c.send, c.entry)
	controls := container.NewHBox(layout.NewSpacer(), c.reasoning)
	bottom := container.NewVBox(controls, input)
	return container.NewBorder(nil, bottom, nil, nil, c.scroll)
}

// ask seeds the input with text and sends it, used for the launcher's initial
// question.
func (c *chatUI) ask(text string) {
	c.entry.SetText(text)
	c.submit()
}

// stop cancels any reply still streaming for this chat. Called when its tab is
// closed so the model is not left generating into a discarded conversation.
func (c *chatUI) stop() {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}

// syncControls seeds this tab's reasoning toggle from the saved default and
// shows it only for Local AI.
func (c *chatUI) syncControls() {
	if c.reasoning == nil {
		return
	}
	p := fyne.CurrentApp().Preferences()
	c.reasoning.SetChecked(p.Bool(prefLocalThinking))
	if p.StringWithFallback(prefProvider, ProviderClaude) == ProviderLocal {
		c.reasoning.Show()
	} else {
		c.reasoning.Hide()
	}
}

// submit sends the current input to the model and streams the reply back.
func (c *chatUI) submit() {
	text := strings.TrimSpace(c.entry.Text)
	if text == "" || c.busy {
		return
	}

	c.entry.SetText("")
	c.addMessage(wmtheme.UserIcon, text)
	c.history = append(c.history, llms.TextParts(llms.ChatMessageTypeHuman, text))

	if c.titleSetter != nil {
		c.titleSetter(launcher.TruncatePrompt(text))
		c.titleSetter = nil
	}

	cfg := loadConfig()
	if cfg.provider == ProviderLocal && c.reasoning != nil {
		cfg.thinking = c.reasoning.Checked
	}
	llm, err := cfg.newLLM()
	if err != nil {
		c.addMessage(theme.WarningIcon(), err.Error())
		return
	}

	c.setBusy(true)

	// Make the request cancellable so closing the tab (stop) abandons the reply
	// and stops the model mid-flight rather than leaving it to run on.
	hist := c.history
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	// The model call performs network IO - keep it off the render thread and
	// marshal only the UI updates back with fyne.Do. The whole thing is guarded:
	// this module runs inside the compositor process, so a panic here (a bad
	// stream from a local server, a render fault) must not crash the desktop.
	go func() {
		defer recoverAI("chat goroutine")
		defer cancel()

		hist = c.converse(ctx, llm, hist)

		fyne.Do(func() {
			defer recoverAI("chat turn end")
			// Record the exchange on the UI thread (the same goroutine as the
			// human-turn append) so history has a single writer.
			c.history = hist

			c.setBusy(false)
			if c.isActive == nil || c.isActive() { // don't grab focus from another tab
				c.win.Canvas().Focus(c.entry)
			}
		})
	}()
}

// converse runs the model until it answers in words rather than command calls.
// Each command call is run and its result handed back, so it can say what it did.
func (c *chatUI) converse(ctx context.Context, llm llms.Model, hist []llms.MessageContent) []llms.MessageContent {
	tools := desktopTools()
	if len(tools) == 0 && usedTools(hist) {
		// This conversation has already run a command, and a provider rejects a
		// tool result it has no matching tool for. So turning tool use off ih this conversation.
		tools = commandTools()
	}

	for turn := 0; ; turn++ {
		reply := c.addPendingReply()
		final, resp, genErr := c.generate(ctx, llm, hist, tools, reply)
		call := firstToolCall(resp)
		c.finishReply(reply, final, resp, genErr, call != nil)

		if genErr != nil || call == nil {
			if genErr == nil && final != "" {
				hist = append(hist, llms.TextParts(llms.ChatMessageTypeAI, final))
			}
			return hist
		}

		// A command was asked for. The request is recorded as an assistant turn of
		// its own - anything the model said alongside it is shown but not kept.
		hist = append(hist, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{*call},
		})

		line, result := runTool(call.FunctionCall.Name, call.FunctionCall.Arguments)
		c.showCommand(line, result)
		hist = append(hist, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: call.ID,
				Name:       call.FunctionCall.Name,
				Content:    result,
			}},
		})

		if turn+1 >= maxToolTurns {
			c.showNote("Stopped after " + strconv.Itoa(maxToolTurns) + " commands.")
			return hist
		}
	}
}

// generate asks the model for one reply, streaming it into the given row as it arrives.
// It returns the complete text alongside the raw response, which carries any tool call or error.
func (c *chatUI) generate(ctx context.Context, llm llms.Model, hist []llms.MessageContent,
	tools []llms.Tool, reply *pendingReply,
) (string, *llms.ContentResponse, error) {
	var sb strings.Builder

	opts := []llms.CallOption{
		// Without an explicit limit langchaingo defaults max_tokens to 2048,
		// which truncates longer replies mid-sentence (stop_reason max_tokens).
		llms.WithMaxTokens(8192),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			text := sb.String()
			// A local server can emit empty leading chunks before any tokens.
			// Keep the activity indicator up until real text actually arrives.
			// A tool call being spelled out as text counts as not yet arrived.
			if text == "" || looksLikeToolCall(text) {
				return nil
			}
			fyne.Do(func() {
				defer recoverAI("chat stream render")
				reply.stop() // first real reply text has arrived
				// Plain Markdown while streaming, hold off on code copy blocks until complete.
				// The copy buttons are added once the reply is complete (final render below).
				reply.body.ParseMarkdown(text)
				c.scroll.ScrollToBottom()
			})
			return nil
		}),
	}
	if len(tools) > 0 {
		opts = append(opts, llms.WithTools(tools))
	}

	resp, err := llm.GenerateContent(ctx, hist, opts...)
	if err != nil {
		return "", nil, err
	}

	final := sb.String()
	if final == "" && resp != nil && len(resp.Choices) > 0 {
		final = resp.Choices[0].Content // provider streamed nothing; use the full reply
	}
	return final, resp, nil
}

// renderMarkdown parses Markdown into the reply body and upgrades its code
// blocks to copyable ones, keeping the two steps together at every render site.
func renderMarkdown(body *widget.RichText, md string) {
	body.ParseMarkdown(md)
	withCopyableCode(body)
}

// addMessage appends a transcript row (an icon beside a Markdown body) and
// returns the body so a streaming reply can update it in place.
func (c *chatUI) addMessage(icon fyne.Resource, text string) *widget.RichText {
	body := widget.NewRichTextFromMarkdown(text)
	body.Wrapping = fyne.TextWrapWord

	row := container.NewBorder(nil, nil, container.NewVBox(widget.NewIcon(icon), layout.NewSpacer()), nil, body)
	c.log.Add(container.NewPadded(row))
	c.log.Refresh()
	c.scroll.ScrollToBottom()
	return body
}

// pendingReply is the assistant's row in the transcript while a reply is in flight.
type pendingReply struct {
	body *widget.RichText
	row  fyne.CanvasObject
	stop func()
}

// addPendingReply appends the assistant's reply row with a running activity
// indicator beside the empty body.
func (c *chatUI) addPendingReply() *pendingReply {
	var reply *pendingReply
	fyne.DoAndWait(func() {
		body := widget.NewRichTextFromMarkdown("")
		body.Wrapping = fyne.TextWrapWord

		spinner := widget.NewActivity()
		spinner.Start()

		waiting := container.NewHBox(spinner)
		content := container.NewVBox(waiting, body)
		row := container.NewPadded(container.NewBorder(nil, nil,
			container.NewVBox(widget.NewIcon(Icon), layout.NewSpacer()), nil, content))
		c.log.Add(row)
		c.log.Refresh()
		c.scroll.ScrollToBottom()

		stopped := false
		reply = &pendingReply{body: body, row: row, stop: func() {
			if stopped {
				return
			}
			stopped = true
			spinner.Stop()
			content.Remove(waiting)
		}}
	})
	return reply
}

// finishReply renders a completed turn into its row - the model's words or error.
// A turn that only asked to run a command is empty so comes out of the transcript.
func (c *chatUI) finishReply(reply *pendingReply, final string, resp *llms.ContentResponse,
	genErr error, called bool,
) {
	truncated := genErr == nil && resp != nil && len(resp.Choices) > 0 &&
		resp.Choices[0].StopReason == "max_tokens"

	fyne.DoAndWait(func() {
		reply.stop() // ensure it stops even if nothing streamed
		switch {
		case genErr != nil:
			reply.body.ParseMarkdown("⚠️ " + genErr.Error())
		case called:
			c.log.Remove(reply.row)
			c.log.Refresh()
		case final != "":
			shown := final
			if truncated {
				shown += " …\n\n*(reply truncated - ask to continue)*"
			}
			renderMarkdown(reply.body, shown)
		default:
			c.log.Remove(reply.row) // nothing said at all: leave no empty row
			c.log.Refresh()
		}
		c.scroll.ScrollToBottom()
	})
}

// showCommand notes a command the model ran, so the person reading can see what
// the assistant did to their desktop rather than only its account of it.
func (c *chatUI) showCommand(line, result string) {
	text := "tyde_ctl " + line
	if problem, failed := strings.CutPrefix(result, "error: "); failed {
		if line == "" {
			text = problem // nothing ran, so there is no command line to show
		} else {
			text += " - " + problem
		}
	}
	c.addNote(text, fyne.TextStyle{Monospace: true})
}

// showNote adds a line from the desktop rather than the model, such as a limit
// having been reached.
func (c *chatUI) showNote(text string) {
	c.addNote(text, fyne.TextStyle{Italic: true})
}

// addNote appends a quiet line from the desktop rather than from either party
// in the conversation.
func (c *chatUI) addNote(text string, style fyne.TextStyle) {
	fyne.DoAndWait(func() {
		note := widget.NewRichText(&widget.TextSegment{Text: text, Style: widget.RichTextStyle{
			ColorName: theme.ColorNameDisabled,
			SizeName:  theme.SizeNameCaptionText,
			TextStyle: style,
		}})
		note.Wrapping = fyne.TextWrapWord

		// Indent past where the speaker icons go, so the note lines up with what
		// was said rather than hanging off the edge of the conversation.
		indent := canvas.NewRectangle(color.Transparent)
		indent.SetMinSize(fyne.NewSquareSize(theme.IconInlineSize()))

		c.log.Add(container.NewPadded(container.NewBorder(nil, nil, indent, nil, note)))
		c.log.Refresh()
		c.scroll.ScrollToBottom()
	})
}

// setBusy toggles the input while a reply is in flight.
func (c *chatUI) setBusy(busy bool) {
	c.busy = busy
	if busy {
		c.send.Disable()
	} else {
		c.send.Enable()
	}
}

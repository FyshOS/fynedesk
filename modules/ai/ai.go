// Package ai adds an AI assistant to Tyde. It is a launch-suggestion module:
// typing in the launcher offers an "Ask AI" option at the end of the
// suggestions which opens a chat window backed by an LLM. Anthropic (Claude)
// and OpenAI are cloud providers reached with an API key; "Local AI" talks to
// an OpenAI-compatible server running on the machine (ollama, the Kronk model
// server, LM Studio, llama.cpp's server, ...) so nothing but the desktop's own
// HTTP request leaves the box. All three go through the langchaingo library and
// are selected and configured from the module's settings panel.
//
// The module is off by default - enable it and pick a provider (adding a key
// for the cloud ones, or a server URL for Local AI) from the "Set up AI" step
// of the first-run setup, or from the AI tab of the settings window.
package ai

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyshos.com/tyde/modules/launcher"

	"fyshos.com/tyde"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

// ModuleName is the registered name of the AI assistant module. It is exported
// so the desktop can identify the module (e.g. to order its launcher
// suggestion after the web search item).
const ModuleName = "AI Assistant"

// Supported providers. Claude and OpenAI are cloud services reached over the
// network with an API key; Local AI is an OpenAI-compatible server the user
// runs on the machine, so it needs a base URL instead of a key.
const (
	ProviderClaude = "Claude"
	ProviderOpenAI = "OpenAI"
	ProviderLocal  = "Local AI"
)

// Preference keys. Cloud keys are stored per-provider so switching provider
// does not discard the other provider's token; Local AI stores its server URL.
// Model overrides are per-provider for the same reason - and because a model
// name only means anything to the provider it was chosen for.
const (
	prefProvider      = "ai.provider"
	prefClaudeKey     = "ai.key.claude"
	prefOpenAIKey     = "ai.key.openai"
	prefLocalEndpoint = "ai.endpoint.local"
	prefTools         = "ai.tools"          // let the assistant run desktop commands (default on)
	prefLocalThinking = "ai.local.thinking" // let a reasoning model think (slower)
	prefLocalManaged  = "ai.local.managed"  // let tyde run the server via kronk (default on)
	prefClaudeModel   = "ai.model.claude"
	prefOpenAIModel   = "ai.model.openai"
	prefLocalModel    = "ai.model.local"
)

// modelPref maps a provider to the preference holding its optional model
// override.
func modelPref(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return prefOpenAIModel
	case ProviderLocal:
		return prefLocalModel
	default:
		return prefClaudeModel
	}
}

// Default models for the cloud providers. Claude defaults to Sonnet 4.6 rather
// than a newer Opus/Sonnet 5 model because langchaingo (v0.1.14) always sends a
// "temperature" field, which the Opus 4.7+/Sonnet 5 family reject with a 400 -
// Sonnet 4.6 still accepts it. Users can override either in settings.
const (
	defaultClaudeModel = "claude-sonnet-4-6"
	defaultOpenAIModel = "gpt-4o"
)

// Local AI depends on which server we expect. With the kronk binary present
// tyde runs and owns a kronk server (fixed port 11435; models named by Hugging
// Face GGUF id) - so there is no URL to configure, just a model. Otherwise we
// assume the user runs their own server, defaulting to ollama (port 11434; its
// own model names) but allowing any OpenAI-compatible server via the Base URL.
// Endpoints include the "/v1" prefix because langchaingo appends
// "/chat/completions".
const (
	kronkEndpoint  = "http://localhost:11435/v1"
	kronkModel     = "unsloth/Qwen3-0.6B-Q8_0"
	ollamaEndpoint = "http://localhost:11434/v1"
	ollamaModel    = "llama3.2"
)

// defaultLocalModel picks the kronk or ollama model so a beginner with either
// installed works with no configuration.
func defaultLocalModel() string {
	if kronkManaged() {
		return kronkModel
	}
	return ollamaModel
}

//go:embed assistant.svg
var assistantSvg []byte

// Icon is the assistant's themed icon, exported for reuse in the settings and
// setup UIs so the module and its configuration screens share one glyph.
var Icon fyne.Resource = theme.NewThemedResource(&fyne.StaticResource{
	StaticName:    "assistant.svg",
	StaticContent: assistantSvg,
})

var meta = tyde.ModuleMetadata{
	Name:        ModuleName,
	NewInstance: newAI,
}

// assistant is the module instance. It owns a single chat window (reused across
// launches, hidden rather than destroyed) holding a tab per conversation.
type assistant struct {
	win   fyne.Window
	tabs  *container.DocTabs
	chats map[*container.TabItem]*chatUI
}

func newAI() tyde.Module {
	// If Local AI is the active provider, start a managed kronk server now (when
	// the module loads) so it is ready by the time the user asks something. A
	// no-op unless kronk is installed and nothing already answers.
	if cfg := loadConfig(); cfg.provider == ProviderLocal {
		serverMgr.ensure()
	}
	return &assistant{}
}

func (a *assistant) Metadata() tyde.ModuleMetadata {
	return meta
}

func (a *assistant) Destroy() {
	// The module cache is torn down and rebuilt on any settings change, not just
	// when the assistant is switched off - so only stop the managed local server
	// when the module is genuinely no longer enabled, otherwise an unrelated
	// tweak would needlessly bounce (and re-download into) the server.
	if !moduleEnabled() {
		serverMgr.stop()
	}
	if a.win != nil {
		a.win.Close()
		a.win = nil
	}
}

// moduleEnabled reports whether the AI Assistant is still in the user's enabled
// module list (preference "modulenames", "|"-separated - the same store the
// settings screen writes). Read at Destroy time to tell a real deactivation
// from a routine module-cache rebuild.
func moduleEnabled() bool {
	names := fyne.CurrentApp().Preferences().String("modulenames")
	for _, n := range strings.Split(names, "|") {
		if n == ModuleName {
			return true
		}
	}
	return false
}

// LaunchSuggestions offers an "Ask AI" item whenever the launcher has input and
// an API key is configured - without a key the assistant can't answer, so the
// suggestion is hidden until it's set up.
func (a *assistant) LaunchSuggestions(input string) []tyde.LaunchSuggestion {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	if !loadConfig().ready() {
		return nil
	}
	return []tyde.LaunchSuggestion{&launchItem{a: a, text: input}}
}

// Shortcuts binds the assistant to the dedicated assistant ("Copilot") key.
func (a *assistant) Shortcuts() map[*tyde.Shortcut]func() {
	return map[*tyde.Shortcut]func(){
		tyde.NewShortcut("Show AI Assistant", tyde.KeyAssistant,
			fyne.KeyModifierShift|fyne.KeyModifierSuper): a.ShowChat,
	}
}

// ShowChat opens the assistant with no question set, as the "Fathom" launcher
// (tyde_ctl fathom) does - the chat window with an empty conversation ready for
// input. Must be called on the UI thread.
func (a *assistant) ShowChat() {
	a.open("")
}

// open shows the chat window (creating it on first use) and, if given, opens the
// question in a new tab. Must be called on the UI thread - it is, via the
// launcher button tap.
func (a *assistant) open(prompt string) {
	if a.win == nil {
		a.chats = map[*container.TabItem]*chatUI{}
		a.tabs = container.NewDocTabs()
		a.tabs.CreateTab = func() *container.TabItem {
			var item *container.TabItem
			item = a.newTab("", func(s string) {
				item.Text = s
				a.tabs.Refresh()
			})
			return item
		}
		a.tabs.OnClosed = func(item *container.TabItem) {
			if chat, ok := a.chats[item]; ok {
				chat.stop()
				delete(a.chats, item)
			}
		}

		watermark := canvas.NewImageFromResource(Icon)
		watermark.FillMode = canvas.ImageFillContain
		watermark.Translucency = 0.9
		watermark.SetMinSize(fyne.NewSquareSize(240))

		entry := widget.NewMultiLineEntry()
		startNew := func() {
			a.open(entry.Text)
			entry.SetText("")
		}
		entry.SetPlaceHolder("Ask fathom...")
		entry.Wrapping = fyne.TextWrapWord
		entry.OnSubmitted = func(_ string) { startNew() }

		send := &widget.Button{
			Icon: theme.MailSendIcon(), Importance: widget.HighImportance,
			OnTapped: startNew,
		}
		input := container.NewBorder(nil, nil, nil, send, entry)

		win := fyne.CurrentApp().NewWindow("Fathom AI")
		win.SetIcon(Icon)
		win.SetContent(container.NewStack(container.NewBorder(nil, input, nil, nil, container.NewCenter(watermark)), a.tabs))
		win.Resize(fyne.NewSize(460, 560))
		win.SetCloseIntercept(win.Hide) // reuse the window (hide, don't destroy)
		a.win = win
	}

	switch {
	case strings.TrimSpace(prompt) != "":
		a.addTab(a.newTab(prompt, nil)) // a new search term opens its own conversation
	case len(a.tabs.Items) == 0:
		var item *container.TabItem
		item = a.newTab("", func(s string) {
			item.Text = s
			a.tabs.Refresh()
		})

		a.addTab(item) // opened bare with nothing yet - give an empty chat
	}

	a.win.Show()
	a.win.RequestFocus()
}

// addTab appends a tab and makes it the visible one.
func (a *assistant) addTab(item *container.TabItem) {
	a.tabs.Append(item)
	a.tabs.Select(item)
}

// newTab builds a fresh conversation as a tab. With a prompt it titles the tab
// from it and asks straight away; empty it is a blank "New chat" ready for
// input.
func (a *assistant) newTab(prompt string, titleSetter func(string)) *container.TabItem {
	chat := newChatUI()
	chat.win = a.win
	chat.titleSetter = titleSetter

	title := "New chat"
	if p := strings.TrimSpace(prompt); p != "" {
		title = launcher.TruncatePrompt(p)
	}
	item := container.NewTabItem(title, chat.build())

	chat.isActive = func() bool { return a.tabs != nil && a.tabs.Selected() == item }
	a.chats[item] = chat

	chat.syncControls() // reflect current provider and reasoning setting
	if strings.TrimSpace(prompt) != "" {
		chat.ask(prompt)
	}
	return item
}

// launchItem is the "Ask AI" launcher suggestion.
type launchItem struct {
	a    *assistant
	text string
}

func (i *launchItem) Icon() fyne.Resource { return Icon }

func (i *launchItem) Title() string { return "Ask AI: " + launcher.TruncatePrompt(i.text) }

func (i *launchItem) Launch() { i.a.open(i.text) }

// config is a snapshot of the persisted AI settings.
type config struct {
	provider string
	model    string // optional override; empty means use the provider default
	key      string // API key for a cloud provider
	endpoint string // server base URL for Local AI
	thinking bool   // Local AI: let a reasoning model think before answering
}

// loadConfig reads the current AI configuration from app preferences.
func loadConfig() config {
	p := fyne.CurrentApp().Preferences()
	prov := p.StringWithFallback(prefProvider, ProviderClaude)

	c := config{provider: prov, model: strings.TrimSpace(p.String(modelPref(prov)))}
	switch prov {
	case ProviderOpenAI:
		c.key = p.String(prefOpenAIKey)
	case ProviderLocal:
		// A managed kronk server lives at a fixed address we own; only a
		// user-run server needs a configurable URL.
		if kronkManaged() {
			c.endpoint = kronkEndpoint
		} else {
			c.endpoint = strings.TrimSpace(p.StringWithFallback(prefLocalEndpoint, ollamaEndpoint))
		}
		c.thinking = p.Bool(prefLocalThinking)
	default:
		c.key = p.String(prefClaudeKey)
	}
	return c
}

// ready reports whether the configured provider can answer: cloud providers
// need an API key, Local AI needs a server URL (which defaults, so it is
// effectively always ready). It gates the "Ask AI" launcher suggestion.
func (c config) ready() bool {
	if c.provider == ProviderLocal {
		return c.endpoint != ""
	}
	return strings.TrimSpace(c.key) != ""
}

// modelName resolves the model to use, falling back to the provider default.
func (c config) modelName() string {
	if c.model != "" {
		return c.model
	}
	return defaultModel(c.provider)
}

// defaultModel is the model a provider uses when the user has not named one.
func defaultModel(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return defaultOpenAIModel
	case ProviderLocal:
		return defaultLocalModel()
	default:
		return defaultClaudeModel
	}
}

// newLLM builds a langchaingo model for the configured provider, or an error
// describing what the user still needs to set up.
func (c config) newLLM() (llms.Model, error) {
	if !c.ready() {
		if c.provider == ProviderLocal {
			return nil, fmt.Errorf("no Local AI server URL set - add one in the AI settings")
		}
		return nil, fmt.Errorf("no %s API key set - add one in the AI settings", c.provider)
	}

	switch c.provider {
	case ProviderOpenAI:
		return openai.New(openai.WithToken(c.key), openai.WithModel(c.modelName()))
	case ProviderLocal:
		// A local OpenAI-compatible server (ollama, Kronk, ...). It ignores the
		// token but langchaingo requires a non-empty one, so pass a placeholder.
		// The injectDoer sets "enable_thinking" per the user's Reasoning toggle:
		// off (default) makes a reasoning model answer directly and fast; on lets
		// it think first - slower but more accurate. Harmless to servers that
		// don't use the field.
		llm, err := openai.New(
			openai.WithBaseURL(c.endpoint),
			openai.WithToken("local"),
			openai.WithModel(c.modelName()),
			openai.WithHTTPClient(injectDoer{
				fields: map[string]any{"enable_thinking": c.thinking},
				next:   http.DefaultClient,
			}),
		)
		if err != nil {
			return nil, err
		}
		// Wrap so a server that isn't up / crashed gives an actionable error.
		return localModel{Model: llm}, nil
	default:
		return anthropic.New(anthropic.WithToken(c.key), anthropic.WithModel(c.modelName()))
	}
}

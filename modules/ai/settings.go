package ai

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// cloudHelp sits beneath the cloud-provider form.
const cloudHelp = "Your key is stored on this device and sent only to the chosen provider.\n" +
	"Leave Model blank to use the provider's default."

// SettingsContent builds the provider configuration panel for the AI assistant.
// Changes are persisted immediately to app preferences so they take effect the
// next time the assistant is opened. The panel is embedded both in the settings
// window's AI tab and in the first-run "Set up AI" step.
//
// Each provider needs different setup - a cloud key, or a local server URL - so
// the area below the provider selector is rebuilt for the chosen provider
// rather than relabelled in place (Fyne forms don't update item labels reliably
// on refresh).
func SettingsContent() fyne.CanvasObject {
	p := fyne.CurrentApp().Preferences()

	provider := widget.NewSelect([]string{ProviderClaude, ProviderOpenAI, ProviderLocal}, nil)
	provider.SetSelected(p.StringWithFallback(prefProvider, ProviderClaude))

	body := container.NewStack()

	var rebuild func()
	rebuild = func() {
		switch provider.Selected {
		case ProviderOpenAI:
			body.Objects = []fyne.CanvasObject{cloudSettings(p, ProviderOpenAI, prefOpenAIKey)}
		case ProviderLocal:
			body.Objects = []fyne.CanvasObject{localSettings(p, rebuild)}
		default:
			body.Objects = []fyne.CanvasObject{cloudSettings(p, ProviderClaude, prefClaudeKey)}
		}
		body.Refresh()
	}
	rebuild()

	provider.OnChanged = func(string) {
		p.SetString(prefProvider, provider.Selected)
		if provider.Selected == ProviderLocal {
			// Switching to Local while the module is loaded: get the managed
			// server coming up so it's ready by the time they ask something.
			serverMgr.ensure()
		}
		rebuild()
	}

	// Tool use is a property of the assistant rather than of a provider, so it
	// sits below the provider's own setup and survives switching between them.
	tools := widget.NewCheck("Allow Fathom to run desktop commands", func(on bool) {
		p.SetBool(prefTools, on)
	})
	tools.SetChecked(p.BoolWithFallback(prefTools, true))

	head := widget.NewForm(widget.NewFormItem("Provider", provider))
	return container.NewVBox(head, body, tools)
}

// cloudSettings builds the API-key + model form for a cloud provider. Both the
// key and the model are stored under that provider's own preference, so
// switching provider leaves the other one's setup untouched.
func cloudSettings(p fyne.Preferences, provider, keyPref string) fyne.CanvasObject {
	key := widget.NewPasswordEntry()
	key.SetPlaceHolder("API key / token")
	key.SetText(p.String(keyPref))
	key.OnChanged = func(s string) { p.SetString(keyPref, s) }

	modelKey := modelPref(provider)
	model := widget.NewEntry()
	model.SetText(p.String(modelKey))
	model.SetPlaceHolder(defaultModel(provider))
	model.OnChanged = func(s string) { p.SetString(modelKey, s) }

	help := widget.NewLabel(cloudHelp)
	help.Wrapping = fyne.TextWrapWord

	return container.NewVBox(widget.NewForm(
		widget.NewFormItem("API Key", key),
		widget.NewFormItem("Model", model),
	), help)
}

// localSettings builds the Local AI panel. rebuild swaps the panel between managed
// vs configuration when the managed option is toggled.
func localSettings(p fyne.Preferences, rebuild func()) fyne.CanvasObject {
	model := widget.NewEntry()
	model.SetText(p.String(prefLocalModel))
	model.SetPlaceHolder(defaultLocalModel())
	model.OnChanged = func(s string) { p.SetString(prefLocalModel, s) }

	status := widget.NewLabel("")
	status.Wrapping = fyne.TextWrapWord

	managed := widget.NewCheck(managedLabel, nil)
	managed.SetChecked(kronkManaged())
	managed.OnChanged = func(on bool) {
		p.SetBool(prefLocalManaged, on)
		if on {
			serverMgr.ensure()
		} else {
			// Hand the server back to the user: stop the one we started, as from
			// here on tyde issues no kronk commands.
			serverMgr.stop()
		}
		rebuild()
	}
	if !kronkAvailable() {
		managed.Disable() // nothing to manage with
	}

	reasoning := widget.NewCheck("Reasoning by default: slower, but more accurate", nil)
	reasoning.SetChecked(p.Bool(prefLocalThinking))
	reasoning.OnChanged = func(on bool) { p.SetBool(prefLocalThinking, on) }

	if kronkManaged() {
		// Managed server: tyde knows the address, so the user only picks a model.
		test := probeButton(func() string { return kronkEndpoint }, model, status)
		return container.NewVBox(
			managed,
			widget.NewForm(widget.NewFormItem("Model", model)),
			reasoning, test, status,
		)
	}

	// User-run server: needs an address.
	endpoint := widget.NewEntry()
	endpoint.SetText(p.StringWithFallback(prefLocalEndpoint, ollamaEndpoint))
	endpoint.SetPlaceHolder(ollamaEndpoint)
	endpoint.OnChanged = func(s string) { p.SetString(prefLocalEndpoint, s) }

	test := probeButton(func() string {
		if ep := strings.TrimSpace(endpoint.Text); ep != "" {
			return ep
		}
		return ollamaEndpoint
	}, model, status)

	return container.NewVBox(
		managed,
		widget.NewForm(
			widget.NewFormItem("Ollama URL", endpoint),
			widget.NewFormItem("Model", model),
		),
		reasoning, test, status,
	)
}

// managedLabel offers tyde running the local server itself, when kronk is there
// to run it with.
const managedLabel = "Automatically manage with Kronk"

// Test-connection outcome marks. A pass needs no words, so it is just the tick
// beside the button; a failure adds the reason below.
const (
	passMark = "✅"
	failMark = "❌"
)

// probeButton makes a "Test connection" button that checks the server at
// endpointFn() and marks the result beside itself, filling in status only when
// something is wrong.
func probeButton(endpointFn func() string, model *widget.Entry, status *widget.Label) fyne.CanvasObject {
	mark := widget.NewLabel("")

	btn := widget.NewButton("Test connection", func() {
		ep := endpointFn()
		want := strings.TrimSpace(model.Text)
		if want == "" {
			want = defaultLocalModel()
		}
		mark.SetText("")
		status.SetText("Testing …")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			res := probeEndpoint(ctx, ep)
			problem := describeProbe(res, want)
			fyne.Do(func() {
				if problem == "" {
					mark.SetText(passMark)
				} else {
					mark.SetText(failMark)
				}
				status.SetText(problem)
			})
		}()
	})

	return container.NewHBox(btn, mark)
}

// describeProbe reduces a connection test to what went wrong, in one line a
// beginner can act on - or "" when the server and model are both ready, since
// the tick says that on its own.
func describeProbe(res probeResult, model string) string {
	if res.err != nil {
		// Unwrap the http.Client's "Get \"url\": …" wrapper: the reason is the
		// useful half and the URL is on screen already.
		err := res.err
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return err.Error()
	}
	if _, ok := res.has(model); ok {
		return ""
	}
	if len(res.models) == 0 {
		return "No models loaded yet - pull '" + model + "' on the server, or wait for its first-use download."
	}
	return "'" + model + "' isn't available. The server offers: " + strings.Join(res.models, ", ")
}

package wm

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"os/user"
	"sync"

	"fyshos.com/fynedesk"
	wmTheme "fyshos.com/fynedesk/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	deskDriver "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/godbus/dbus/v5"
)

type subj struct {
	Kind    string
	Details map[string]dbus.Variant
}

type auth struct {
	windows map[string]fyne.Window
}

func (a *auth) register() {
	conn2, _ := dbus.SystemBus()
	err := conn2.ExportAll(a, "/AuthenticationAgent", "org.freedesktop.PolicyKit1.AuthenticationAgent")
	if err != nil {
		fyne.LogError("Could not start auth agent server", err)
	}

	obj := conn2.Object("org.freedesktop.PolicyKit1", "/org/freedesktop/PolicyKit1/Authority")
	call := obj.Call("org.freedesktop.PolicyKit1.Authority.RegisterAuthenticationAgent", 0,

		&subj{"unix-session", map[string]dbus.Variant{
			"session-id": dbus.MakeVariant("c1"),
		}}, "en_US",
		"/AuthenticationAgent")
	if call.Err != nil {
		fyne.LogError("Failed to register auth agent", call.Err)
	}
}

type ident struct {
	ID      string
	Details map[string]dbus.Variant
}

func (a *auth) BeginAuthentication(actionID, message, iconName string, details map[string]string, cookie string, ids []ident, sender dbus.Sender) (err *dbus.Error) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	pass := widget.NewPasswordEntry()

	username := ""
	uid := ""
	_, err2 := fmt.Sscanf(ids[0].Details["uid"].String(), "@u %s", &uid)
	if err2 != nil {
		currentUser, err2 := user.Current()
		if err2 == nil {
			username = currentUser.Username
		} else {
			fyne.LogError("Failed to look up fallback user", err2)
		}
	} else {
		usr, err2 := user.LookupId(uid)
		if err2 == nil {
			username = usr.Username
		} else {
			fyne.LogError("Failed to look up user "+uid, err2)
		}
	}
	f := widget.NewForm(
		widget.NewFormItem("Ident", widget.NewLabel(username)),
		widget.NewFormItem("Password", pass),
	)
	w := fyne.CurrentApp().Driver().(deskDriver.Driver).CreateSplashWindow()
	a.windows[cookie] = w

	var auth *widget.Button
	auth = widget.NewButton("Authorize", func() {
		auth.Disable()
		cmd := exec.Command("/usr/lib/polkit-1/polkit-agent-helper-1", username)

		buffer := bytes.Buffer{}
		buffer.Write([]byte(cookie + "\n"))
		buffer.Write([]byte(pass.Text + "\n"))
		cmd.Stdin = &buffer

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err3 := cmd.Run()

		if err3 != nil {
			log.Println("Auth err", err3)
		} else {
			w.Close()
		}
		auth.Enable()
	})
	auth.Importance = widget.HighImportance
	cancel := widget.NewButton("Cancel", func() {
		w.Close()
	})
	pass.OnSubmitted = func(string) {
		auth.OnTapped()
	}

	header := widget.NewRichTextFromMarkdown(fmt.Sprintf("### Authorise\n\n_%s_", message))
	header.Truncation = fyne.TextTruncateEllipsis
	bottomPad := canvas.NewRectangle(color.Transparent)
	bottomPad.SetMinSize(fyne.NewSquareSize(10))
	content := container.NewBorder(
		header,
		container.NewVBox(
			container.NewHBox(layout.NewSpacer(),
				container.NewGridWithColumns(2, cancel, auth),
				layout.NewSpacer()), bottomPad),
		nil, nil, f)

	r, g, b, _ := theme.OverlayBackgroundColor().RGBA()
	bgCol := &color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 230}

	bg := canvas.NewRectangle(bgCol)
	icon := canvas.NewImageFromResource(wmTheme.LockIcon)
	iconBox := container.NewWithoutLayout(icon)
	icon.Resize(fyne.NewSize(92, 92))
	icon.Move(fyne.NewPos(300-92-theme.Padding(), theme.Padding()))
	w.SetContent(container.NewStack(
		iconBox, bg,
		container.NewPadded(content)))

	w.SetOnClosed(func() {
		delete(a.windows, cookie)
		wg.Done()
	})
	fynedesk.Instance().WindowManager().ShowModal(w, fyne.NewSize(300, 210))

	wg.Wait()
	return nil
}

func (a *auth) CancelAuthentication(cookie string, sender dbus.Sender) (err *dbus.Error) {
	if w, ok := a.windows[cookie]; ok {
		w.Close() // OnClose will tidy the session
	}
	return nil
}

// StartAuthAgent asks our policy kit agent to start listening for auth requests.
func StartAuthAgent() {
	a := &auth{windows: make(map[string]fyne.Window)}
	go a.register()
}

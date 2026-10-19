package ui

import (
	"context"
	"errors"
	"fmt"
	"image"

	// Register decoders so any common image can be re-encoded as the .face PNG.
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty"

	wmtheme "fyshos.com/tyde/theme"
)

// facePath returns the location of the current user's avatar image. This matches
// the convention read by the fin login manager (see ../fin newAvatar).
func facePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".face")
}

// loadAccountScreen builds the Account tab: the user's avatar (stored at
// ~/.face) and a change-password form driving the system passwd command.
func (d *settingsUI) loadAccountScreen() fyne.CanvasObject {
	name := "user"
	if u, err := user.Current(); err == nil {
		if u.Name != "" {
			name = u.Name
		} else {
			name = u.Username
		}
	}

	return container.NewVScroll(container.NewVBox(
		d.loadUserImageCard(),
		widget.NewSeparator(),
		d.loadPasswordCard(name),
		widget.NewSeparator(),
		d.loadFingerprintCard(),
	))
}

// loadFingerprintCard builds the fingerprint enrolment and management UI.
func (d *settingsUI) loadFingerprintCard() fyne.CanvasObject {
	client, err := newFprintClient()
	if err != nil {
		return container.NewVBox(widget.NewRichTextFromMarkdown("## Fingerprint"),
			widget.NewLabel("No fingerprint sensor detected."))
	}
	d.fprint = client // closed when the settings window closes

	enrolledBox := container.NewVBox()
	d.refreshEnrolled(enrolledBox)

	fingerSelect := widget.NewSelect(nil, nil)
	for _, f := range fingerNames {
		fingerSelect.Options = append(fingerSelect.Options, fingerLabel(f))
	}
	fingerSelect.SetSelectedIndex(0)

	enroll := widget.NewButtonWithIcon("Enrol", theme.ContentAddIcon(), func() {
		d.startEnroll(fingerNames[fingerSelect.SelectedIndex()], enrolledBox)
	})
	enroll.Importance = widget.HighImportance
	enrollRow := container.NewBorder(nil, nil, widget.NewLabel("Add a finger:"), enroll, fingerSelect)

	content := container.NewVBox(
		enrolledBox,
		widget.NewSeparator(),
		enrollRow,
		widget.NewSeparator(),
		d.fingerprintLoginToggle(),
	)
	return container.NewVBox(
		widget.NewRichTextFromMarkdown("## Fingerprint\n\nUnlock the screen and log in with a fingerprint"),
		content,
	)
}

// refreshEnrolled repopulates the list of enrolled fingers, each with a delete button.
func (d *settingsUI) refreshEnrolled(box *fyne.Container) {
	box.Objects = nil
	fingers, err := d.fprint.listEnrolled()
	if err != nil || len(fingers) == 0 {
		box.Add(widget.NewLabel("No fingerprints enrolled."))
		box.Refresh()
		return
	}
	for _, f := range fingers {
		row := container.NewBorder(nil, nil,
			widget.NewIcon(theme.ConfirmIcon()),
			widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				d.deleteEnrolled(box)
			}),
			widget.NewLabel(fingerLabel(f)))
		box.Add(row)
	}
	box.Refresh()
}

// deleteEnrolled removes all enrolled fingerprints after confirmation.
func (d *settingsUI) deleteEnrolled(box *fyne.Container) {
	dialog.ShowConfirm("Remove fingerprints",
		"Remove all enrolled fingerprints for your account?", func(ok bool) {
			if !ok {
				return
			}
			go func() {
				err := d.fprint.deleteEnrolled()
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, d.win)
					}
					d.refreshEnrolled(box)
				})
			}()
		}, d.win)
}

// startEnroll runs an interactive enrollment for the chosen finger, showing live
// scan progress in a dialog.
func (d *settingsUI) startEnroll(finger string, enrolledBox *fyne.Container) {
	status := widget.NewLabel("Touch the sensor with your " + fingerLabel(finger) + "...")
	status.Wrapping = fyne.TextWrapWord
	bar := widget.NewProgressBar()
	prog := dialog.NewCustom("Enrolling "+fingerLabel(finger), "Cancel",
		container.NewVBox(status, bar), d.win)
	prog.Show()

	done := make(chan struct{})
	prog.SetOnClosed(func() {
		select {
		case <-done:
		default:
			// Cancelled mid-enrolment: stop the sensor and refresh the list.
			d.fprint.device.Call(fprintDeviceIf+".EnrollStop", 0)
			d.fprint.device.Call(fprintDeviceIf+".Release", 0)
		}
	})

	go func() {
		err := d.fprint.enroll(finger, func(msg string, frac float64) {
			fyne.Do(func() {
				status.SetText(msg)
				bar.SetValue(frac)
			})
		})
		close(done)
		fyne.Do(func() {
			prog.Hide()
			if err != nil {
				dialog.ShowError(err, d.win)
			}
			d.refreshEnrolled(enrolledBox)
		})
	}()
}

func (d *settingsUI) fingerprintLoginToggle() fyne.CanvasObject {
	check := widget.NewCheck("Use fingerprint to log in and unlock the screen", nil)
	check.SetChecked(fingerprintLoginEnabled())
	check.OnChanged = func(on bool) {
		check.Disable()
		go func() {
			err := setFingerprintLogin(on)
			fyne.Do(func() {
				check.Enable()
				if err != nil {
					dialog.ShowError(err, d.win)
					check.SetChecked(fingerprintLoginEnabled()) // revert to reality
				}
			})
		}()
	}

	if !pamFprintdAvailable() {
		warn := widget.NewLabel("Install the pam_fprintd module to enable this.")
		warn.Importance = widget.WarningImportance
		return container.NewVBox(check, warn)
	}
	return check
}

// loadUserImageCard shows the current avatar and lets the user pick a new one,
// which is re-encoded to a PNG at ~/.face so the login manager picks it up too.
func (d *settingsUI) loadUserImageCard() fyne.CanvasObject {
	avatar := canvas.NewImageFromResource(wmtheme.UserIcon)
	if p := facePath(); p != "" {
		if _, err := os.Stat(p); err == nil {
			avatar = canvas.NewImageFromFile(p)
		}
	}
	avatar.FillMode = canvas.ImageFillContain
	avatar.SetMinSize(fyne.NewSize(112, 112))

	pickDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil || file == nil {
			return
		}
		defer func() { _ = file.Close() }()

		if err := writeUserFace(file); err != nil {
			dialog.ShowError(err, d.win)
			return
		}

		// Force a reload from disk now that the file has changed.
		avatar.Resource = nil
		avatar.File = facePath()
		avatar.Refresh()

		// Reflect the new picture on the widget panel's account button too.
		if d.panel != nil {
			d.panel.refreshAccountIcon()
		}
	}, d.win)
	pickDialog.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".gif"}))
	if dir, err := getPicturesDir(); err == nil {
		pickDialog.SetLocation(dir)
	}

	change := widget.NewButtonWithIcon("Change Image...", theme.FolderOpenIcon(), func() {
		pickDialog.Show()
	})

	return container.NewBorder(nil, nil, container.NewCenter(avatar), nil,
		container.NewVBox(
			widget.NewLabel("Your picture is shown on the login screen."),
			container.NewHBox(change, layout.NewSpacer()),
		))
}

// loadPasswordCard builds the change-password form.
func (d *settingsUI) loadPasswordCard(name string) fyne.CanvasObject {
	current := widget.NewPasswordEntry()
	current.SetPlaceHolder("Current password")
	next := widget.NewPasswordEntry()
	next.SetPlaceHolder("New password")
	confirm := widget.NewPasswordEntry()
	confirm.SetPlaceHolder("Confirm new password")

	currentItem := widget.NewFormItem("Current", current)
	currentItem.Required = true
	nextItem := widget.NewFormItem("New", next)
	nextItem.Required = true
	confirmItem := widget.NewFormItem("Confirm", confirm)
	confirmItem.Required = true

	form := &widget.Form{
		Items:      []*widget.FormItem{currentItem, nextItem, confirmItem},
		SubmitText: "Change Password",
		Validator: func() error {
			if next.Text != confirm.Text {
				return errors.New("new passwords do not match")
			}
			return nil
		},
	}
	form.OnSubmit = func() {
		form.Disable()
		old, updated := current.Text, next.Text
		// Driving passwd blocks on IO, so keep it off the render thread.
		go func() {
			err := changePassword(old, updated)
			fyne.Do(func() {
				form.Enable()
				if err != nil {
					dialog.ShowError(err, d.win)
					return
				}
				current.SetText("")
				next.SetText("")
				confirm.SetText("")
				dialog.ShowInformation("Password Changed",
					"Your password was updated successfully.", d.win)
			})
		}()
	}

	return container.NewVBox(
		widget.NewRichTextFromMarkdown("## Change Password\n\nSigned in as "+name),
		form,
	)
}

// writeUserFace decodes the chosen image and re-encodes it as a PNG at ~/.face.
func writeUserFace(src io.Reader) error {
	img, _, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("unsupported image: %w", err)
	}

	dst := facePath()
	if dst == "" {
		return errors.New("could not resolve home directory")
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if err := png.Encode(out, img); err != nil {
		return err
	}
	return out.Close()
}

// changePassword drives the system passwd command through a pty, feeding the
// current then new password. passwd reads via getpass() which needs a terminal,
// so a plain stdin pipe is not enough - hence the pty.
func changePassword(oldPass, newPass string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "passwd")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("could not start passwd: %w", err)
	}
	defer func() { _ = ptmx.Close() }()

	// The prompts arrive in order: current password, then the new password
	// twice. Send the current one first and the new one for anything after.
	var transcript strings.Builder
	sent := 0
	buf := make([]byte, 1024)
	for {
		n, readErr := ptmx.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			transcript.WriteString(chunk)
			if sent < 3 && strings.Contains(strings.ToLower(chunk), "password:") {
				reply := newPass
				if sent == 0 {
					reply = oldPass
				}
				if _, err := io.WriteString(ptmx, reply+"\n"); err != nil {
					break
				}
				sent++
			}
		}
		if readErr != nil { // EOF once passwd exits and closes the pty
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errors.New("changing password timed out")
		}
		return errors.New(passwdError(transcript.String()))
	}
	return nil
}

// passwdError extracts the most useful line from passwd's output, falling back
// to a generic message when nothing recognisable was printed.
func passwdError(out string) string {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		low := strings.ToLower(line)
		if strings.Contains(low, "authentication") || strings.Contains(low, "bad password") ||
			strings.Contains(low, "sorry") || strings.Contains(low, "too short") ||
			strings.Contains(low, "not changed") {
			return strings.TrimPrefix(line, "passwd: ")
		}
	}
	return "password was not changed (incorrect current password?)"
}

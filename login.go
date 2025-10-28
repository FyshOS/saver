package saver

import (
	"image/color"
	"os/user"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var loginDialog dialog.Dialog

func showLogin(unlocked func(), w fyne.Window) {
	if loginDialog != nil {
		return
	}

	showCursor(w)
	input := newPasswordEscapeEntry(func() {
		if loginDialog != nil {
			loginDialog.Hide()
		}
	})
	user, _ := user.Current()

	tryUnlock := func() {
		spin := widget.NewActivity()
		prop := canvas.NewRectangle(color.Transparent)
		prop.SetMinSize(fyne.NewSquareSize(96))
		d := dialog.NewCustomWithoutButtons("Unlocking...",
			container.NewStack(prop, spin), w)

		spin.Start()
		d.Show()

		go func() {
			defer func() {
				fyne.Do(func() {
					d.Hide()
					spin.Stop()
				})
			}()

			if !canUnlock(user.Username, input.Text) {
				fyne.Do(func() {
					hideCursor(w)
				})
				return
			}

			fyne.Do(unlocked)
		}()
	}
	dismiss := func() {
		w.Canvas().Focus(nil)
		loginDialog = nil
	}

	input.OnSubmitted = func(_ string) {
		loginDialog.Hide()
		showCursor(w) // dialog dismissing hid it
		tryUnlock()
	}
	w.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		switch ev.Name {
		case fyne.KeyEscape:
			if loginDialog != nil {
				loginDialog.Hide()
			}
		}
	})

	loginDialog = dialog.NewForm("Enter Password", "Unlock", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Username", widget.NewLabel(user.Username)),
			widget.NewFormItem("Password", input),
		},
		func(ok bool) {
			dismiss()
			if !ok {
				hideCursor(w)
				return
			}

			tryUnlock()
		}, w)
	loginDialog.Resize(fyne.NewSize(340, 90))
	loginDialog.Show()
	go func() {
		time.Sleep(time.Millisecond * 100)
		fyne.Do(func() {
			w.Canvas().Focus(input)
		})
	}()
}

type passwordEscapeEntry struct {
	widget.Entry

	esc func()
}

func newPasswordEscapeEntry(fn func()) *passwordEscapeEntry {
	p := &passwordEscapeEntry{esc: fn}
	p.ExtendBaseWidget(p)

	p.Password = true
	return p
}

func (p *passwordEscapeEntry) TypedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeyEscape {
		p.esc()
		return
	}

	p.Entry.TypedKey(ev)
}

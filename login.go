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
	input := widget.NewPasswordEntry()
	user, _ := user.Current()

	tryUnlock := func() {
		spin := widget.NewActivity()
		prop := canvas.NewRectangle(color.Transparent)
		prop.SetMinSize(fyne.NewSquareSize(96))
		d := dialog.NewCustomWithoutButtons("Unlocking...",
			container.NewStack(prop, spin), w)

		spin.Start()
		d.Show()
		defer func() {
			d.Hide()
			spin.Stop()
		}()

		if !canUnlock(user.Username, input.Text) {
			return
		}

		unlocked()
	}
	dismiss := func() {
		hideCursor(w)
		w.Canvas().Focus(nil)
		loginDialog = nil
	}

	input.OnSubmitted = func(_ string) {
		loginDialog.Hide()
		dismiss()
		tryUnlock()
	}

	loginDialog = dialog.NewForm("Enter Password", "Unlock", "Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Username", widget.NewLabel(user.Username)),
			widget.NewFormItem("Password", input),
		},
		func(ok bool) {
			dismiss()
			if !ok {
				return
			}

			tryUnlock()
		}, w)
	loginDialog.Resize(fyne.NewSize(340, 90))
	loginDialog.Show()
	go func() {
		time.Sleep(time.Millisecond * 100)
		w.Canvas().Focus(input)
	}()
}

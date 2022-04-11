package main

import (
	"bytes"
	"dkeymask/core"
	"dkeymask/theme"
	"net/url"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func main() {
	app := app.New()
	app.Settings().SetTheme(&theme.CustomTheme{})
	window := app.NewWindow("DKeyMask - a better way to keep secret")
	windowSize := fyne.NewSize(480, 720)
	entrydataSize := fyne.NewSize(480, 480)
	window.SetIcon(theme.WindowIcon())
	window.Resize(windowSize)
	window.SetFixedSize(true)
	window.SetPadded(false)

	entrydata := widget.NewMultiLineEntry()
	entrydata.Wrapping = fyne.TextWrapBreak
	entrydata.SetPlaceHolder("secret data here...\n\nusage:\n\t[READ] read secret data from an image(png/jpg)\n\t[WRITE] write secret data to an image(png/jpg)\n\nsuggestion:\n\t- use passphrase for secondary protection(AES)\n\t- keep the network offline when using\n\t- rename the output image file manually\n\nprompt:\n\t- 1MB image file can store about less than 100kb data\n\t- don't compress the result image, which will cause data loss\n\t- source https://github.com/geekmask/dkeymask\n\t- 100% OPEN & SAFE & FREE, star if you like :)")
	entrydata.Resize(entrydataSize)
	entrydatac := container.New(layout.NewGridWrapLayout(entrydataSize), entrydata)
	entrypwd := widget.NewEntry()
	entrypwd.SetPlaceHolder("option passphrase...")
	infile := widget.NewLabel("infile:")
	outfile := widget.NewLabel("outfile:")
	url, _ := url.Parse("https://github.com/geekmask/dkeymask")
	about := widget.NewHyperlink("donate: dgeek.eth", url)

	infile.Wrapping = fyne.TextWrapBreak
	infile.TextStyle = fyne.TextStyle{Italic: true}
	outfile.Wrapping = fyne.TextWrapBreak
	outfile.TextStyle = fyne.TextStyle{Italic: true}
	about.Alignment = fyne.TextAlignCenter

	infile.Hide()
	outfile.Hide()

	btnr := widget.NewButton("READ", func() {
		if len(entrypwd.Text) > 32 {
			infile.Text = "infile: passphrase exceeds limit 32bytes!"
			infile.Refresh()
			infile.Show()
			outfile.Hide()
			return
		}
		openFile(window, func(r fyne.URIReadCloser, err error) {
			if r != nil {
				infile.Text = "infile: " + r.URI().Path()
				infile.Refresh()
				infile.Show()
				outfile.Hide()

				f, _ := os.Open(r.URI().Path())
				defer f.Close()

				resultBytes, err := core.Decode(f)
				if err != nil {
					infile.Text = "infile: " + err.Error()
					infile.Refresh()
					infile.Show()
					return
				}
				if len(entrypwd.Text) > 0 {
					resultBytes, err = core.AESDecrypt(resultBytes, []byte(entrypwd.Text))
					if err != nil {
						infile.Text = "infile: " + err.Error()
						infile.Refresh()
						infile.Show()
						return
					}
					entrypwd.SetText("")
				}
				entrydata.SetText(string(resultBytes))
			}
		})
	})
	btnw := widget.NewButton("WRITE", func() {
		if len(entrydata.Text) == 0 {
			infile.Text = "infile: the secret data can't be empty!"
			infile.Refresh()
			infile.Show()
			outfile.Hide()
			return
		}
		if len(entrypwd.Text) > 32 {
			infile.Text = "infile: passphrase exceeds limit 32bytes!"
			infile.Refresh()
			infile.Show()
			outfile.Hide()
			return
		}
		openFile(window, func(r fyne.URIReadCloser, err error) {
			if r != nil {
				infile.Text = "infile: " + r.URI().Path()
				infile.Refresh()
				infile.Show()

				f, _ := os.Open(r.URI().Path())
				defer f.Close()
				dir, _ := filepath.Split(r.URI().Path())
				f2, _ := os.Create(dir + "result.png")
				defer f2.Close()
				resultBytes := []byte(entrydata.Text)
				if len(entrypwd.Text) > 0 {
					resultBytes, err = core.AESEncrypt(resultBytes, []byte(entrypwd.Text))
					if err != nil {
						infile.Text = "infile: " + err.Error()
						return
					}
					entrypwd.SetText("")
				}
				err = core.Encode(f, bytes.NewReader(resultBytes), f2)
				if err == nil {
					p, _ := filepath.Abs(f2.Name())
					outfile.Text = "outfile: " + p
				} else {
					outfile.Text = "outfile: " + err.Error()
				}
				outfile.Refresh()
				outfile.Show()
				entrydata.SetText("")
			}
		})
	})

	window.SetContent(container.NewVBox(entrypwd, entrydatac, infile, outfile, layout.NewSpacer(), container.NewGridWithColumns(2, btnr, btnw), about))
	window.ShowAndRun()
}

func openFile(window fyne.Window, callback func(fyne.URIReadCloser, error)) {
	fileDialog := dialog.NewFileOpen(callback, window)
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
	fileDialog.Resize(window.Canvas().Size())
	fileDialog.Show()
}

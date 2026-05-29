package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Run() {
	a := app.New()
	w := a.NewWindow("nosmec")

	topBar := widget.NewLabel("nosmec")
	topBar.Alignment = fyne.TextAlignCenter

	sidebar := container.NewVBox(
		widget.NewButton("Timeline", nil),
		widget.NewButton("Profile", nil),
		widget.NewButton("Relay List", nil),
		widget.NewButton("Settings", nil),
	)
	sidebarBorder := container.NewBorder(nil, nil, nil, nil, sidebar)

	content := widget.NewLabel("Welcome to nosmec")
	content.Alignment = fyne.TextAlignCenter

	split := container.NewHSplit(sidebarBorder, content)
	split.SetOffset(0.25)

	topContainer := container.NewVBox(topBar, split)

	w.SetContent(topContainer)
	w.Resize(fyne.NewSize(1024, 768))
	w.ShowAndRun()
}
package main

import (
	"ShortcutVisualizer/internal/ui"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	w := ui.NewMainWindow(a)

	w.Window().Resize(fyne.NewSize(600, 600))
	w.Window().ShowAndRun()
}

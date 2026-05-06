package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.New()
	w := a.NewWindow("Triangle - 图像识别自动化工具")
	w.Resize(fyne.NewSize(800, 650))
	w.SetFixedSize(false)

	triangle := NewTriangleApp(a, w)
	w.SetContent(triangle.Build())

	w.ShowAndRun()
}

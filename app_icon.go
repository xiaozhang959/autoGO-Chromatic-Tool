package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed build/logo.png
var appIconData []byte

var appIconResource fyne.Resource = fyne.NewStaticResource("logo.png", appIconData)

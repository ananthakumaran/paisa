package web

import (
	"embed"
)

//go:embed static/*
var Static embed.FS

//go:embed static/index.html
var Index string

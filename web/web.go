package web

import (
	"embed"
)

//go:embed all:static/*
var Static embed.FS

//go:embed static/index.html
var Index string

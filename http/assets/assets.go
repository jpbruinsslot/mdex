package assets

import (
	"embed"
)

//go:embed css/*
//go:embed img/*
var FS embed.FS

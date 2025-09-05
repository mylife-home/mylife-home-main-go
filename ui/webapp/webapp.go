package webapp

import (
	"embed"
)

//go:embed all:dist
var FS embed.FS

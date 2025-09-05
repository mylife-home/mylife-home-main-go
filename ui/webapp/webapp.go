package webapp

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var build embed.FS

var FS fs.FS

func init() {
	var err error
	FS, err = fs.Sub(build, "dist/prod/static")

	if err != nil {
		panic(err)
	}
}

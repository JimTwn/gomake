package main

import (
	_ "embed"
	"html/template"
)

var assetTemplate = template.Must(template.New("build").Parse(`package main

import "log"

func main() {
{{range $fn := .Functions}}
	log.Println("> {{$fn}}")
	{{$fn}}()
{{end}}
}
`))

const assetModFile = `
module gomake

go 1.21.5

require (
	github.com/jimtwn/gomake v0.1.0
)
`

// This program should include a copy of the gomake library code.
// We will be adding this to any build.go program we generate, so that
// doesn't have to fetch a copy from Github every time build is run.

//go:embed build/main.go
var assetLibraryMainGo []byte

//go:embed go.mod
var assetLibraryGoMod []byte

const assetLibraryModulesTxt = `
# github.com/jimtwn/gomake v0.1.0
## explicit; go 1.21.5
`

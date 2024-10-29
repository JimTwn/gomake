package main

import (
	_ "embed"
	"fmt"
	"html/template"
)

var assetTemplate = template.Must(template.New("build").Parse(`package main

func main() {
{{range $fn := .Functions}}
	{{$fn}}()
{{end}}
}
`))

var assetModFile = fmt.Sprintf(`
module gomake

go 1.21.5

require (
	github.com/jimtwn/gomake %s
)
`, AppVersion)

// This program should include a copy of the gomake library code.
// We will be adding this to any build.go program we generate, so that
// doesn't have to fetch a copy from Github every time build is run.

//go:embed build/main.go
var assetLibraryMainGo []byte

//go:embed go.mod
var assetLibraryGoMod []byte

var assetLibraryModulesTxt = fmt.Sprintf(`
# github.com/jimtwn/gomake %s
## explicit; go 1.21.5
`, AppVersion)

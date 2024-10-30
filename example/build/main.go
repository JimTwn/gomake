package main

import (
	"log"
	"os"

	gm "github.com/jimtwn/gomake"
)

func main() {
	// Define the 'foo' and 'bar' build units and their dependencies.
	// They are marked as 'default', so they will run if this program
	// is invoked without expicitely mentioning any units.
	assets := gm.AddUnit("assets", installAssets, true)
	foo := gm.AddUnit("foo", buildFoo, true)
	foo.AddDependency(assets)

	// The 'clean' unit does not depend on anything.
	// This one is not marked as 'default', so it will not run if this
	// program is invoked without expicitely mentioning any units.
	_ = gm.AddUnit("clean", clean, false)

	gm.MkdirAll("bin", 0700)
	gm.Build(os.Args[1:])
}

// Copy assets to bin/ directory.
func installAssets() {
	log.Println("> installAssets")
	gm.CopyDir("bin", "assets")
}

// Compile foo and move it to the bin/ directory.
func buildFoo() {
	log.Println("> buildFoo")
	gm.GoBuild("foo", "-o", gm.Join("..", "bin", "foo.exe"))
}

func clean() {
	log.Println("> clean")
	gm.DeleteDirs("bin")
	gm.GoClean(".", "./..")
}

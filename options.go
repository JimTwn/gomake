package main

import (
	"flag"
	"fmt"
	"os"

	b "github.com/jimtwn/gomake/build"
)

type Options struct {
	Root      string
	Rules     []string
	OutputDir string
}

// ParseOptions parses commandline arguments and returns build options.
func ParseOptions() Options {
	opt := newOptions()

	flag.Usage = func() {
		w := os.Stderr
		exeName, err := os.Executable()
		if err != nil {
			b.Throw("os.Executable: %v", err)
		}

		fmt.Fprintf(w, "%s [options] [<build root>]\n\n", exeName)
		flag.PrintDefaults()
	}

	help := flag.Bool("help", false, "Print this help.")
	version := flag.Bool("version", false, "Prints version information.")
	flag.StringVar(&opt.Root, "C", ".", "Build root directory.")
	flag.StringVar(&opt.OutputDir, "out", opt.OutputDir, "Output directory, relative to the build root.")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *version {
		fmt.Fprintf(os.Stderr, "%s %s, %s", AppName, AppVersion, AppVendor)
		os.Exit(0)
	}

	opt.Rules = flag.Args()
	opt.Root = b.Abs(opt.Root)
	return opt
}

func newOptions() Options {
	cwd, err := os.Getwd()
	if err != nil {
		b.Throw("os.Getwd: %v", err)
	}

	return Options{
		Root:      cwd,
		OutputDir: "bin",
	}
}

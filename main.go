package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"runtime"
	"strings"

	b "github.com/jimtwn/gomake/build"
)

func main() {
	log.Default().SetFlags(0)
	log.Default().SetPrefix("")

	// Create a temporary build dir.
	buildDir := b.MkdirTemp("", "gomake*")

	// Intercept any panics.
	//
	// We want to reach this top level before exiting from an error,
	// so we can cleanly remove the build directory we created above.
	// `defer f()` alone is not enough if a fatal error has occurred
	// as any os.Exit calls eliminate pending defers.
	defer func() {
		x := recover()
		b.DeleteDirs(buildDir)

		if x != nil {
			fmt.Fprintln(os.Stderr, x)
			os.Exit(1)
		}
	}()

	run(buildDir)
}

func run(buildDir string) {
	options := ParseOptions()

	// Make root the current working directory. Incase it isn't already.
	b.Chdir(options.Root)

	// Ensure the `build.go` file is sane and the requested rules exist.
	options.Rules = verifyBuildFile(options.Rules)

	// Copy the user provided `build.go`` file to buildDir.
	b.CopyFile(b.Join(buildDir, "build.go"), "build.go")

	// Run the build program template to generate a `main.go` file.
	// Its `main()` function will call the build rules in the user specified `build.go`.
	runBuildTemplate(buildDir, options.Rules)

	// Install a copy of our gomake library.
	installModule(buildDir)

	// Run `go build` in the buildDir to generate our custom build program.
	builderName := getBuilderName()
	b.GoBuild(buildDir, "-o", builderName)

	// Ensure the output dir exists.
	b.MkdirAll(options.OutputDir, 0700)

	// Finally, run our newly built program in the current directory.
	b.Run(b.Join(buildDir, builderName))
}

// getBuilderName determines the build program file name.
// This is platform specific.
func getBuilderName() string {
	if strings.EqualFold(runtime.GOOS, "windows") {
		return "builder.exe"
	}
	return "builder"
}

// The custom build program needs access to our gomake library.
// We have to manually install it next to the build program source.
func installModule(dir string) {
	{
		fd := b.CreateFile(b.Join(dir, "go.mod"))
		defer fd.Close()
		fd.Write([]byte(assetModFile))
	}

	// Install the gomake library.
	//
	// We copy our embedded version of the library into the vendor directory
	// instead of calling `go get`. This ensures the whole thing is self-contained.
	//
	// We have to include a sane go.mod and vendor/modules.txt file to ensure
	// `go build` works as intended.

	vendorDir := b.Join(dir, "vendor", "github.com", "jimtwn", "gomake")
	vendorLibraryDir := b.Join(vendorDir, "build")

	b.MkdirAll(vendorLibraryDir, 0700)
	b.CopyFileData(b.Join(dir, "vendor", "modules.txt"), []byte(assetLibraryModulesTxt))
	b.CopyFileData(b.Join(vendorDir, "go.mod"), assetLibraryGoMod)
	b.CopyFileData(b.Join(vendorLibraryDir, "main.go"), assetLibraryMainGo)
}

// runBuildTemplate generates a `main.go` file which calls the user provided build rules.
func runBuildTemplate(dir string, rules []string) {
	functions := translateRuleNames(rules...)
	fd := b.CreateFile(b.Join(dir, "main.go"))
	defer fd.Close()

	if err := assetTemplate.Execute(fd, struct {
		Functions []string
	}{
		Functions: functions,
	}); err != nil {
		b.Throw("template.Execute: %v", err)
	}
}

// verifyBuildFile parses the  `build.go` source and ensures it is sane.
// It additionalyl checks of the file contains all functions reflecting the
// build rules invoked by the user.
//
// If @rules is empty, this finds any defined BuildXXX functions with
// `//gomake:default` decorators and uses those instead.
func verifyBuildFile(rules []string) []string {
	// Check if there is a `build.go` file in the current directory.
	if !b.FileExists("build.go") {
		b.Throw("no build.go file found or file is not valid")
	}

	// Parse the source.
	fset := token.NewFileSet()
	src, err := parser.ParseFile(fset, "build.go", nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		b.Throw("parser.ParseFile: %v", err)
	}

	// Check if all rules are defined in the file.
	if len(rules) > 0 {
		for _, rule := range rules {
			if !containsFunction(src, rule) {
				b.Throw("build.go does not define a function matching the rule %q", rule)
			}
		}
		return rules
	}

	// No rules are defined. Use the functions defined in the source.
	out := make([]string, 0, len(src.Decls))

	for _, d := range src.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			// Does the function have the expected name?
			if !strings.HasPrefix(fn.Name.Name, "Build") && len(fn.Name.Name) > 5 {
				continue
			}

			// Does the function have the `//gomake:default` decorator?
			if !hasDefaultDecorator(fn.Doc.List) {
				continue
			}

			out = append(out, strings.ToLower(fn.Name.Name[5:]))
		}
	}

	return out
}

// hasDefaultDecorator retursn true if @list contains a comment representing
// the `//gomake:default` decorator.
func hasDefaultDecorator(list []*ast.Comment) bool {
	for _, v := range list {
		if strings.Contains(v.Text, "//gomake:default") {
			return true
		}
	}
	return false
}

// containsFunction returns true if @src contains a function for a rule with the given @name.
func containsFunction(src *ast.File, rule string) bool {
	for _, d := range src.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "Build") && strings.EqualFold(fn.Name.Name[5:], rule) {
				return true
			}
		}
	}
	return false
}

// translateRuleNames returs a new list where each entry from @rules is
// translated into a build function name. These follow the form:
//
//   - "clean" => "BuildClean"
//   - "install" => "BuildInstall"
func translateRuleNames(rules ...string) []string {
	rules = filterStrings(rules)
	out := make([]string, len(rules))

	for i, rule := range rules {
		out[i] = "Build" + strings.ToTitle(rule[:1]) + rule[1:]
	}

	return out
}

// filterStrings returns @set, minus any empty or duplicate entries.
func filterStrings(set []string) []string {
	out := make([]string, 0, len(set))
	for _, str := range set {
		str = strings.TrimSpace(str)
		if len(str) == 0 || containsString(out, str) {
			continue
		}
		out = append(out, str)
	}
	return out
}

// containsString returns true if @set contains @v.
func containsString(set []string, v string) bool {
	for _, str := range set {
		if str == v {
			return true
		}
	}
	return false
}

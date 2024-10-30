package gomake

import "runtime"

// BuildTarget defines a compilation target.
type BuildTarget struct {
	OS   string
	Arch string
}

// BuildOptions defines some build properties set by the user or the system.
type BuildOptions struct {
	Target BuildTarget
}

// NewOptions returns a default build option set.
func NewBuildOptions() BuildOptions {
	return BuildOptions{
		Target: BuildTarget{
			OS:   findGoEnv("GOOS", runtime.GOOS),
			Arch: findGoEnv("GOARCH", runtime.GOARCH),
		},
	}
}

// returns the value for the given Go environment variable.
// Unlike GoEnv(), this checks multiple sources:
//
//  1. System environment - in case thie current command was invoked with custom ENV settings.
//  2. `go env` for the current go environment setting.
//  3. if neither of the above apply, returns @fallback.
func findGoEnv(key string, fallback string) string {
	v, ok := SysEnv(key)
	if !ok {
		v, ok = GoEnv(key)
		if !ok {
			v = fallback
		}
	}
	return v
}

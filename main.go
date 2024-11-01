package gomake

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Build runs all build units defined in the given set (usually commandline arguments.)
// If no explicit unit is specified, this runs all units marked as default.
func Build(units []string) {
	unit_lock.Lock()
	defer unit_lock.Unlock()

	if len(units) == 0 {
		units = getDefaultUnits()
	}

	for _, unit := range units {
		runUnit(unit)
	}
}

// getDefaultUnits returns the names for all units marked "default".
func getDefaultUnits() []string {
	out := make([]string, 0, len(unit_cache))

	for name, unit := range unit_cache {
		if unit.IsDefault() {
			out = append(out, name)
		}
	}

	return out
}

// runUnit runs the unit by the given name, if it exists and is not already finished.
// This additionaly runs the units this unit depends on.
func runUnit(name string) {
	name = strings.ToLower(name)
	unit, ok := unit_cache[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown unit name %q\n", name)
		os.Exit(1)
	}
	unit.Run()
}

// FileExists returns true if the given path exists and is a regular file.
func FileExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		Throw("os.Stat: %v", err)
	}
	return !stat.IsDir()
}

// DirExists returns true if the given path exists and is a directory.
func DirExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		Throw("os.Stat: %v", err)
	}
	return stat.IsDir()
}

// Join joins any number of path elements into a single path,
// separating them with an OS specific Separator. Empty elements
// are ignored. The result is Cleaned. However, if the argument
// list is empty or all its elements are empty, Join returns
// an empty string.
//
// On Windows, the result will only be a UNC path if the first
// non-empty element is a UNC path.
func Join(elems ...string) string {
	return filepath.Join(elems...)
}

// panics with the given formatted error.
func Throw(msg string, args ...any) {
	panic(fmt.Errorf(msg, args...))
}

// Abs returns an absolute representation of path.
//
// If the path is not absolute it will be joined with the current
// working directory to turn it into an absolute path. The absolute
// path name for a given file is not guaranteed to be unique.
// Abs calls Clean on the result.
func Abs(path string) string {
	v, err := filepath.Abs(path)
	if err != nil {
		Throw("Abs: %v", err)
	}
	return v
}

// DeleteFiles deletes the given files.
// The paths are expected to be relative to the build root.
func DeleteFiles(files ...string) {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			Throw("DeleteFiles: %v", err)
		}
	}
}

// DeleteDirs recursively deletes the given directories and all their contents.
// The paths are expected to be relative to the build root.
func DeleteDirs(dirs ...string) {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			Throw("DeleteDirs: %v", err)
		}
	}
}

// Chdir changes the current working directory to the named directory.
func Chdir(path string) {
	if err := os.Chdir(path); err != nil {
		Throw("Chdir: %v", err)
	}
}

// MkdirTemp creates a new temporary directory in the directory dir
// and returns the pathname of the new directory. The new directory's
// name is generated by adding a random string to the end of pattern.
//
// If pattern includes a "*", the random string replaces the last
// "*" instead. If dir is the empty string, MkdirTemp uses the
// default directory for temporary files, as returned by TempDir.
//
// Multiple programs or goroutines calling MkdirTemp simultaneously
// will not choose the same directory. It is the caller's responsibility
// to remove the directory when it is no longer needed.
func MkdirTemp(dir string, pattern string) string {
	dir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		Throw("MkdirTemp: %v", err)
	}
	return dir
}

// Mkdir creates a new directory with the specified name and permission bits (before umask).
// If there is an error, it will be of type *PathError.
func Mkdir(name string, perm os.FileMode) {
	if err := os.Mkdir(name, perm); err != nil {
		Throw("Mkdir: %v", err)
	}
}

// MkdirAll creates a directory named path, along with any necessary parents.
// The permission bits perm (before umask) are used for all directories that
// MkdirAll creates. If path is already a directory, MkdirAll does nothing.
func MkdirAll(name string, perm os.FileMode) {
	if err := os.MkdirAll(name, perm); err != nil {
		Throw("MkdirAll: %v", err)
	}
}

// Create creates or truncates the named file.
//
// If the file already exists, it is truncated. If the file does not exist,
// it is created with mode 0666 (before umask). If successful, methods on
// the returned File can be used for I/O; the associated file descriptor
// has mode O_RDWR. If there is an error, it will be of type *PathError.
func CreateFile(path string) *os.File {
	fd, err := os.Create(path)
	if err != nil {
		Throw("CreateFile: %v", err)
	}
	return fd
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY. If there is an error, it will be of
// type *PathError.
func OpenFile(path string) *os.File {
	fd, err := os.Open(path)
	if err != nil {
		Throw("OpenFile: %v", err)
	}
	return fd
}

// ReadFile reads the named file and returns the contents.
func ReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		Throw("ReadFile: %v", err)
	}
	return data
}

// MoveFile moves file @src to @dst.
// This is equivalent to calling:
//
// > CopyFile(dst, src); DeleteFiles(src)
func MoveFile(dst string, src string) {
	CopyFile(dst, src)
	DeleteFiles(src)
}

// CopyDir recursively copies directory @src and all its contents (including sub-dirs) to directory @dst.
func CopyDir(dst string, src string) {
	MkdirAll(dst, 0700)

	entries, err := os.ReadDir(src)
	if err != nil {
		Throw("CopyDir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			CopyDir(Join(dst, entry.Name()), Join(src, entry.Name()))
		} else {
			CopyFile(Join(dst, entry.Name()), Join(src, entry.Name()))
		}
	}
}

// CopyFile copies the contents of file @src to file @dst.
func CopyFile(dst string, src string) {
	fs := OpenFile(src)
	defer fs.Close()

	fd := CreateFile(dst)
	defer fd.Close()

	if _, err := io.Copy(fd, fs); err != nil {
		Throw("CopyFile: %v", err)
	}
}

// CopyFileData copies @src into a file at @dst.
func CopyFileData(dst string, src []byte) {
	fd := CreateFile(dst)
	defer fd.Close()

	if _, err := fd.Write(src); err != nil {
		Throw("CopyFile: %v", err)
	}
}

// LookPath searches for an executable named file in the directories named
// by the PATH environment variable. LookPath also uses PATHEXT environment
// variable to match a suitable candidate. If file contains a slash, it is
// tried directly and the PATH is not consulted. Otherwise, on success, the
// result is an absolute path.
//
// If the file can not be found, this returns @file as-is.
func LookPath(file string) string {
	path, err := exec.LookPath(file)
	if errors.Is(err, exec.ErrNotFound) {
		return file
	}
	if errors.Is(err, exec.ErrDot) {
		return Abs(path)
	}
	return path
}

// GoBuild runs `go build` in the given directory.
// It includes the given additional arguments.
func GoBuild(dir string, args ...string) {
	runGo("GoBuild", "build", dir, args...)
}

// GoInstall runs `go install` in the given directory.
// It includes the given additional arguments.
func GoInstall(dir string, args ...string) {
	runGo("GoInstall", "install", dir, args...)
}

// GoGet runs `go get` in the given directory.
// It includes the given additional arguments.
func GoGet(dir string, args ...string) {
	runGo("GoGet", "get", dir, args...)
}

// GoEnvList runs `go env` and returns all defined variables.
func GoEnvList() map[string]string {
	var stderr, stdout bytes.Buffer
	if !RunRedirected(nil, &stdout, &stderr, true, "go", "env") {
		Throw("GoEnv: %s", stderr.Bytes())
	}

	// Parse the output of `go env`.
	// It comes as a set of lines where each line is of the form: "set KEY=VALUE"
	// The value can be empty.
	out := make(map[string]string)
	scn := bufio.NewScanner(&stdout)
	for scn.Scan() {
		line := strings.TrimSpace(scn.Text())
		if strings.HasPrefix(line, "set ") {
			line = strings.TrimSpace(line[4:])
		}

		if len(line) == 0 {
			continue
		}

		if index := strings.Index(line, "="); index > -1 {
			key := strings.ToUpper(line[:index])
			val := line[index+1:]
			out[key] = val
		}
	}

	if err := scn.Err(); err != nil {
		Throw("GoEnv: %v", err)
	}

	return out
}

// GoEnv returns the value for a specific Go environment variable.
// Returns false if can't be found. Note that the returned string can
// be empty if no value is defined but the key is present.
// E.g.: `"set GO111MODULE="`.
func GoEnv(key string) (string, bool) {
	set := GoEnvList()
	v, ok := set[strings.ToUpper(key)]
	return v, ok
}

// GoGet runs `go mod` in the given directory.
// It includes the given additional arguments.
func GoMod(dir string, args ...string) {
	runGo("GoMod", "mod", dir, args...)
}

// GoTest runs `go test` in the given directory.
// It includes the given additional arguments.
func GoTest(dir string, args ...string) {
	runGo("GoTest", "test", dir, args...)
}

// GoClean runs `go clean` in the given directory.
// It includes the given additional arguments.
func GoClean(dir string, args ...string) {
	runGo("GoClean", "clean", dir, args...)
}

// GoRun runs `go run` in the given directory.
// It includes the given additional arguments.
func GoRun(dir string, args ...string) {
	runGo("GoRun", "run", dir, args...)
}

// Go runs `go` in the given directory.
// It includes the given additional arguments.
func Go(dir string, args ...string) {
	runGoBare("Go", dir, args...)
}

func runGo(caller string, subCommand string, dir string, args ...string) {
	runGoBare(caller, dir, append([]string{subCommand}, args...)...)
}

func runGoBare(caller string, dir string, args ...string) {
	argv := append([]string{"-C", dir}, args...) // -C flag must be first flag on command line
	cmd := exec.Command(LookPath("go"), argv...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		Throw("%s: %v", caller, err)
	}

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			os.Exit(1)
		} else {
			Throw("%s: %v", caller, err)
		}
	}
}

// SysEnv retrieves the value of the system environment variable named by @key.
//
// If the variable is present in the environment, the value (which may be empty)
// is returned and the boolean is true. Otherwise the returned value will be
// empty and the boolean will be false.
//
// Note: this is not the same as `GoEnv`.
func SysEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// Run runs the given command with specified arguments.
// This does not redirect stdin, stdout or stderr. If the
// command returns error, it is treated as a fatal error.
//
// @wait determines if this call should wait for the command to finish or not.
func Run(wait bool, command string, args ...string) {
	var stderr bytes.Buffer
	if !RunRedirected(nil, nil, &stderr, wait, command, args...) {
		Throw("%s", stderr.Bytes())
	}
}

// RunRedirected runs the given command with specified arguments.
// This function allows redirecting of stdin- stdout- and stderr if desired.
// Specify nil for those you are not interested in.
//
// @wait determines if this call should wait for the command to finish or not.
//
// If the executed command generates an error and stderr is specified, this
// call returns false. Returns true otherwise.
func RunRedirected(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	wait bool,
	command string,
	args ...string,
) bool {
	var path string
	if FileExists(command) {
		path = Abs(command)
	} else {
		var err error
		path, err = exec.LookPath(command)
		if err != nil {
			if errors.Is(err, exec.ErrDot) {
				path = Abs(path)
			} else {
				Throw("RunRedirected: %v", err)
			}
		}
	}

	cmd := exec.Command(path, args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Stdin = stdin

	if wait {
		if err := cmd.Run(); err != nil {
			if stderr != nil {
				return false
			}
			Throw("RunRedirected: %v", err)
		}
	} else {
		if err := cmd.Start(); err != nil {
			Throw("RunRedirected: %v", err)
		}
	}

	return true
}

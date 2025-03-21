package proc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

var (
	pid     int
	name    string
	workdir string
	ctx     context.Context
)

func init() {
	var err error
	workdir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	name = filepath.Base(os.Args[0])
	pid = os.Getpid()
	Logger = os.Stdout

	registerSignalListener()
}

// Pid returns pid of the current process.
func Pid() int {
	return pid
}

// Name returns the process name, same as the command name.
func Name() string {
	return name
}

// WorkDir returns working directory of the current process.
func WorkDir() string {
	return workdir
}

// Path returns a path with components of the working directory
func Path(components ...string) string {
	return filepath.Join(workdir, filepath.Join(components...))
}

// Pathf returns a path with format of the working directory.
func Pathf(format string, args ...any) string {
	return filepath.Join(workdir, fmt.Sprintf(format, args...))
}

// Context return the process context.
func Context() context.Context {
	return ctx
}

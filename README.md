# proc — Process utilities for Go

![CI](https://github.com/go-slim/proc/actions/workflows/ci.yml/badge.svg)

[English](README.md) | [简体中文](README_zh.md)

Small, focused helpers for working with the current process, signals, and running child processes.

## Features

- **Process info**: Get process metadata with `Pid()`, `Name()`, `WorkDir()`, `Path(...)`, `Pathf(...)`, `Context()`
- **Signals**: Register listeners with `On()`/`Once()`, remove via `Cancel()`, trigger via `Notify()`
- **Shutdown**: Graceful shutdown with `Shutdown(syscall.Signal)` and configurable force-kill delay (test-friendly via stub)
- **Exec**: Run external commands with timeout, environment variables, working directory, and lifecycle callbacks
- **Logging**: Control debug output via the `Logger` variable

Module path: `go-slim.dev/proc`

## Install

```bash
go get go-slim.dev/proc
```

Go version: `1.24` (per `proc/go.mod`).

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "syscall"
    "time"

    proc "go-slim.dev/proc"
)

func main() {
    // Get process information
    fmt.Println("pid:", proc.Pid(), "name:", proc.Name(), "wd:", proc.WorkDir())

    // Build paths relative to working directory
    configPath := proc.Path("config", "app.yaml")
    fmt.Println("config path:", configPath)

    // Listen for SIGTERM once
    proc.Once(syscall.SIGTERM, func() {
        fmt.Println("terminating...")
    })

    // Run a command with timeout and error handling
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := proc.Exec(ctx, proc.ExecOptions{
        Command: "sh",
        Args:    []string{"-c", "echo ok"},
    })
    if err != nil {
        log.Fatalf("command failed: %v", err)
    }
}
```

## Signals

The signal API allows you to register custom handlers for OS signals:

- **`On(sig, fn) uint32`** - Registers a listener that fires every time the signal is received. Returns a listener ID.
- **`Once(sig, fn) uint32`** - Registers a one-shot listener that automatically removes itself after execution. Returns a listener ID.
- **`Cancel(id...)`** - Removes listeners by their IDs. Safe to call with invalid or already-removed IDs.
- **`Notify(sig) bool`** - Manually triggers callbacks for the given signal. Returns false if no listeners were found.

**Automatic shutdown**: The package installs a signal listener on init for common signals (`SIGHUP`, `SIGINT`, `SIGQUIT`, `SIGTERM`) that triggers graceful shutdown.

### Example: Custom signal handling

```go
import (
    "fmt"
    "syscall"
    proc "go-slim.dev/proc"
)

// Register a repeating handler
id := proc.On(syscall.SIGUSR1, func() {
    fmt.Println("Received SIGUSR1")
})

// Register a one-time handler
proc.Once(syscall.SIGUSR2, func() {
    fmt.Println("This will only run once")
})

// Later: cancel the repeating handler
proc.Cancel(id)
```

## Shutdown

The shutdown system provides graceful termination with configurable force-kill behavior.

```go
import (
    "syscall"
    "time"
    proc "go-slim.dev/proc"
)

// Optional: set a delay before force-kill
proc.SetTimeToForceQuit(2 * time.Second)

// Trigger graceful shutdown sequence
err := proc.Shutdown(syscall.SIGTERM)
if err != nil {
    // handle error
}
```

**Behavior**:
- If `SetTimeToForceQuit()` is called with a duration > 0:
  1. Calls `Notify(SIGTERM)` in a goroutine to trigger registered listeners
  2. Waits for the specified duration
  3. Force-kills the process if still alive
- If delay is 0 or not set:
  1. Calls `Notify(SIGTERM)` synchronously
  2. Immediately kills the process

**Testing**: The `Shutdown` function uses an internal `killFn` variable (defaults to OS kill) which can be stubbed for testing graceful shutdown behavior without actually killing the process.

## Exec

Execute external commands with fine-grained control over timeout, environment, and lifecycle.

```go
import (
    "context"
    "os/exec"
    "time"
    proc "go-slim.dev/proc"
)

ctx := context.Background()
err := proc.Exec(ctx, proc.ExecOptions{
    WorkDir: "/tmp",                           // Working directory (defaults to current)
    Timeout: 3 * time.Second,                  // Maximum execution time
    Env:     []string{"FOO=BAR"},              // Additional environment variables
    Command: "sh",                             // Command to execute
    Args:    []string{"-c", "echo ok"},        // Command arguments
    TTK:     500 * time.Millisecond,           // Time To Kill: grace period between interrupt and force kill
    OnStart: func(cmd *exec.Cmd) {
        fmt.Println("Process started:", cmd.Process.Pid)
    },
})
if err != nil {
    // Handle timeout, cancellation, or execution errors
}
```

### ExecOptions Fields

- **WorkDir**: Working directory for the command (defaults to current process working directory)
- **Timeout**: If > 0, creates a timeout context automatically
- **Env**: Additional environment variables (appended to current process environment)
- **Stdin**, **Stdout**, **Stderr**: Custom I/O streams (defaults to os.Stdout/os.Stderr)
- **Command**: The executable to run
- **Args**: Command-line arguments
- **TTK** (Time To Kill): Delay between sending interrupt signal and kill signal during cancellation
- **OnStart**: Callback invoked after the command starts successfully

### Platform-specific behavior

- **Unix/Linux**: Sets `Setpgid=true` to create a new process group, preventing zombie processes when child processes spawn their own children
- **Windows**: No special process attributes are set

## Logging

Control debug output by setting the `Logger` variable:

```go
import (
    "io"
    "os"
    proc "go-slim.dev/proc"
)

// Disable logging
proc.Logger = io.Discard

// Log to a file
logFile, _ := os.Create("proc.log")
proc.Logger = logFile

// Default: logs to os.Stdout
```

## Use Cases

### Graceful server shutdown

```go
proc.SetTimeToForceQuit(10 * time.Second)
proc.Once(syscall.SIGTERM, func() {
    // Close database connections
    db.Close()
    // Stop accepting new requests
    server.Shutdown(context.Background())
})
```

### Running builds with timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := proc.Exec(ctx, proc.ExecOptions{
    Command: "go",
    Args:    []string{"build", "./..."},
    WorkDir: proc.WorkDir(),
    TTK:     10 * time.Second,
})
```

### Hot reload on signal

```go
proc.On(syscall.SIGHUP, func() {
    // Reload configuration
    config.Reload()
})
```

## Development

Run local CI-like checks:

```bash
make ci
```

Run benchmarks:

```bash
make bench
# or save to artifacts/bench.txt
make bench-save
```

## License

MIT

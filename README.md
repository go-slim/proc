# proc — Process utilities for Go

![CI](https://github.com/go-slim/proc/actions/workflows/ci.yml/badge.svg)

Small, focused helpers for working with the current process, signals, and running child processes.

- Process info: `Pid()`, `Name()`, `WorkDir()`, `Path(...)`, `Pathf(...)`, `Context()`
- Signals: register listeners with `On()`/`Once()`, remove via `Cancel()`, trigger via `Notify()`
- Shutdown: `Shutdown(syscall.Signal)` to gracefully notify then force kill (test-friendly via stub)
- Exec: run external commands with timeout, env, working dir, and `OnStart` callback

Module path: `go-slim.dev/proc`

## Install

```bash
go get go-slim.dev/proc
```

Go version: `1.22` (per `proc/go.mod`).

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "syscall"
    proc "go-slim.dev/proc"
)

func main() {
    fmt.Println("pid:", proc.Pid(), "name:", proc.Name(), "wd:", proc.WorkDir())

    // Listen for SIGTERM once
    proc.Once(syscall.SIGTERM, func() { fmt.Println("terminating...") })

    // Run a short command with timeout
    _ = proc.Exec(context.Background(), proc.ExecOptions{
        Command: "sh",
        Args:    []string{"-c", "echo ok"},
    })
}
```

## Signals

- `On(sig, fn)` registers a listener, `Once(sig, fn)` registers a one-shot listener.
- `Cancel(id...)` removes listeners by id.
- `Notify(sig)` triggers callbacks for `sig`. Returns false if no listeners were found.

The package installs a signal listener on init for common signals (HUP/INT/QUIT/TERM) that gracefully shuts down.

## Shutdown

```go
// Set a delay before force-kill (optional)
proc.SetTimeToForceQuit(2 * time.Second)
// Trigger graceful shutdown sequence
_ = proc.Shutdown(syscall.SIGTERM)
```

- If a delay is set, the package calls `Notify(SIGTERM)`, sleeps, then force-kills.
- For testing, `Shutdown` uses an internal `killFn` variable (defaults to OS kill) which can be stubbed.

## Exec

```go
err := proc.Exec(ctx, proc.ExecOptions{
    WorkDir: "/tmp",
    Timeout: 3 * time.Second,
    Env:     []string{"FOO=BAR"},
    Command: "sh",
    Args:    []string{"-c", "echo ok"},
    TTK:     500 * time.Millisecond, // grace period before SIGKILL after timeout
    OnStart: func(cmd *exec.Cmd) {
        // inspect/adjust cmd before it starts
    },
})
```

Platform notes:
- Unix: `SetSysProcAttribute` enables `Setpgid` to avoid zombie processes.
- Windows: `SetSysProcAttribute` is a no-op.

## Development

- Local CI-like checks:

```bash
make ci
```

- Benchmarks:

```bash
make bench
# or save to artifacts/bench.txt
make bench-save
```

## License

MIT

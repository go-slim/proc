package proc

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// ExecOptions configures command execution parameters.
type ExecOptions struct {
	// WorkDir specifies the working directory for the command.
	// If empty, defaults to the current process's working directory.
	WorkDir string
	// Timeout specifies the maximum duration for command execution.
	// If > 0, a timeout context will be created.
	Timeout time.Duration
	// Env specifies additional environment variables to pass to the command.
	// These are appended to the current process's environment.
	Env []string
	// Stdin specifies the standard input for the command.
	Stdin io.Reader
	// Stdout specifies the standard output for the command.
	Stdout io.Writer
	// Stderr specifies the standard error output for the command.
	Stderr io.Writer
	// Command specifies the command to execute.
	Command string
	// Args specifies the command arguments.
	Args []string
	// TTK (Time To Kill) specifies the delay between sending interrupt signal
	// and kill signal during command cancellation.
	TTK time.Duration
	// OnStart is a callback function invoked after the command starts.
	OnStart func(cmd *exec.Cmd)
}

// Exec executes a command with the given context and options.
// It supports timeout, graceful shutdown with configurable kill delay,
// and proper process group management to prevent zombie processes.
//
// References:
// - https://github.com/gouravkrosx/golang-cmd-exit-demo?ref=hackernoon.com
// - https://keploy.io/blog/technology/managing-go-processes
func Exec(ctx context.Context, opts ExecOptions) error {
	if opts.WorkDir == "" {
		opts.WorkDir = workdir
	}

	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}

	// Run the app as the user who invoked sudo
	// username := os.Getenv("SUDO_USER")
	// cmd := exec.CommandContext(ctx, "sh", "-c", userCmd)
	// if username != "" {
	// 	// print all environment variables
	// 	slog.Debug("env inherited from the cmd", slog.Any("env", os.Environ()))
	// 	// Run the command as the user who invoked sudo to preserve the user environment variables and PATH
	// 	cmd = exec.CommandContext(ctx, "sudo", "-E", "-u", os.Getenv("SUDO_USER"), "env", "PATH="+os.Getenv("PATH"), "sh", "-c", userCmd)
	// }
	cmd := exec.CommandContext(ctx, opts.Command, opts.Args...)
	cmd.Dir = cmp.Or(opts.WorkDir, workdir)
	cmd.Env = append(os.Environ(), opts.Env...)

	// Set the cancel function for the command
	cmd.Cancel = func() error {
		if cancel != nil {
			cancel()
		}
		return nil
	}

	// wait after sending the interrupt signal, before sending the kill signal
	if opts.TTK > 0 {
		cmd.WaitDelay = opts.TTK
	}

	SetSysProcAttribute(cmd)

	// Sets the input of the command
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}

	// Sets the output of the command
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start the app: %w", err)
	}

	if opts.OnStart != nil {
		opts.OnStart(cmd)
	}

	err = cmd.Wait()
	select {
	case <-ctx.Done():
		if ctxerr := ctx.Err(); ctxerr != nil {
			return fmt.Errorf("context cancelled, error while waiting for the app to exit: %w", ctxerr)
		}
		return err
	default:
		if err != nil {
			return fmt.Errorf("unexpected error while waiting for the app to exit: %w", err)
		}
		log.Println("app exited successfully")
		return nil
	}
}

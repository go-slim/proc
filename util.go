package proc

import (
	"fmt"
	"io"
	"os"
)

// Logger is the output destination for debug messages.
// By default, it's set to os.Stdout in the init function.
// Set to io.Discard to disable debug logging.
var Logger io.Writer

// debugf outputs a formatted debug message to the Logger.
// If Logger is nil or io.Discard, no output is produced.
// The format string follows fmt.Printf conventions.
func debugf(format string, args ...any) {
	if Logger != nil && Logger != io.Discard {
		_, err := fmt.Fprintf(Logger, format+"\n", args...)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
		}
	}
}

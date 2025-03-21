package proc

import (
	"fmt"
	"io"
	"os"
)

var Logger io.Writer

func debugf(format string, args ...any) {
	if Logger != nil && Logger != io.Discard {
		_, err := fmt.Fprintf(Logger, format+"\n", args...)
		if err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
		}
	}
}

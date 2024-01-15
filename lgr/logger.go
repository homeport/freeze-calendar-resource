package lgr

import (
	"fmt"
	"io"
)

type LogLevel int

const (
	FatalLevel = LogLevel(iota)
	ErrorLevel
	InfoLevel
	DebugLevel
)

type Logger struct {
	Level  LogLevel
	Writer io.Writer
}

// Write provides compatibility with io.Writer. When using this method, Log levels are ignored and no newline will be added.
func (l Logger) Write(data []byte) (int, error) {
	fmt.Fprint(l.Writer, string(data))
	return 0, nil
}

func (l Logger) Debug(format string, a ...any) {
	if l.Level >= DebugLevel {
		fmt.Fprintf(l.Writer, format, a...)
		fmt.Fprintln(l.Writer)
	}
}

func (l Logger) Info(format string, a ...any) {
	if l.Level >= InfoLevel {
		fmt.Fprintf(l.Writer, format, a...)
		fmt.Fprintln(l.Writer)
	}
}

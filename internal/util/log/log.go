package log

import (
	"fmt"
	"io"
)

type Logger struct {
	Out     io.Writer
	Err     io.Writer
	Verbose bool
	Quiet   bool
}

func New(out io.Writer, err io.Writer, verbose bool, quiet bool) *Logger {
	return &Logger{
		Out:     out,
		Err:     err,
		Verbose: verbose,
		Quiet:   quiet,
	}
}

func (l *Logger) Info(format string, args ...any) {
	if l.Quiet {
		return
	}
	fmt.Fprintf(l.Out, format+"\n", args...)
}

func (l *Logger) Debug(format string, args ...any) {
	if l.Quiet || !l.Verbose {
		return
	}
	fmt.Fprintf(l.Out, format+"\n", args...)
}

func (l *Logger) Error(format string, args ...any) {
	fmt.Fprintf(l.Err, format+"\n", args...)
}

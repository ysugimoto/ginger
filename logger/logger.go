package logger

import (
	"fmt"

	"github.com/mattn/go-colorable"
)

// Ansi colors
const (
	red    = "\033[31m"
	yellow = "\033[93m"
	green  = "\033[92m"
	reset  = "\033[0m"
)

var stdout = colorable.NewColorableStdout()

type Logger struct {
	ns string
}

func WithNamespace(ns string) *Logger {
	return &Logger{
		ns: ns,
	}
}

func (l *Logger) AddNamespace(ns string) {
	l.ns += "." + ns
}

func (l *Logger) Info(message ...interface{}) {
	fmt.Fprintln(stdout, green+"["+l.ns+":INFO]"+fmt.Sprint(message)+reset)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	fmt.Fprintf(stdout, green+"["+l.ns+":INFO]"+format+reset, args...)
}

func (l *Logger) Warn(message ...interface{}) {
	fmt.Fprintln(stdout, yellow+"["+l.ns+":WARN]"+fmt.Sprint(message)+reset)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(stdout, yellow+"["+l.ns+":WARN]"+format+reset, args...)
}

func (l *Logger) Error(message ...interface{}) {
	fmt.Fprintln(stdout, red+"["+l.ns+":ERROR]"+fmt.Sprint(message)+reset)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(stdout, red+"["+l.ns+":ERROR]"+format+reset, args...)
}

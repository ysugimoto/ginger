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

// Logger is the struct that outputs colored log with namespace.
type Logger struct {
	ns string
}

func WithNamespace(ns string) *Logger {
	return &Logger{
		ns: ns,
	}
}

// AddNamespace() adds more namespace string on output log.
func (l *Logger) AddNamespace(ns string) {
	l.ns += "." + ns
}

// Info() outputs information log with green color.
func (l *Logger) Info(message ...interface{}) {
	fmt.Fprintln(stdout, green+"["+l.ns+":INFO] "+fmt.Sprint(message...)+reset)
}

// Infof() outputs formatted information log with green color.
func (l *Logger) Infof(format string, args ...interface{}) {
	fmt.Fprintf(stdout, green+"["+l.ns+":INFO] "+format+reset, args...)
}

// Warn() outputs warning log with yellow color.
func (l *Logger) Warn(message ...interface{}) {
	fmt.Fprintln(stdout, yellow+"["+l.ns+":WARN] "+fmt.Sprint(message...)+reset)
}

// Warnf() outputs formatted warning log with yellow color.
func (l *Logger) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(stdout, yellow+"["+l.ns+":WARN] "+format+reset, args...)
}

// Error() outputs error log with red color.
func (l *Logger) Error(message ...interface{}) {
	fmt.Fprintln(stdout, red+"["+l.ns+":ERROR] "+fmt.Sprint(message...)+reset)
}

// Errorf() outputs formatted error log with red color.
func (l *Logger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(stdout, red+"["+l.ns+":ERROR] "+format+reset, args...)
}

// Print() outputs log with default color.
func (l *Logger) Print(message ...interface{}) {
	fmt.Fprintln(stdout, "["+l.ns+"] "+fmt.Sprint(message...))
}

// Printf() outputs formatted log with default color.
func (l *Logger) Printf(format string, args ...interface{}) {
	fmt.Fprintf(stdout, "["+l.ns+"] "+format, args...)
}

package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger interface defines logging operations
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// StandardLogger implements Logger using logrus
type StandardLogger struct {
	*logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level string, verbose bool) Logger {
	log := logrus.New()
	
	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	
	// Set formatter for colored output
	log.SetFormatter(&ColorFormatter{
		verbose: verbose,
	})
	
	log.SetOutput(os.Stdout)
	
	return &StandardLogger{Logger: log}
}

// ColorFormatter provides colored console output
type ColorFormatter struct {
	verbose bool
}

// Format formats log entries with colors and standardized prefixes
func (f *ColorFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var color string
	var prefix string
	
	switch entry.Level {
	case logrus.DebugLevel:
		color = "\033[36m" // Cyan
		prefix = "[*]"
	case logrus.InfoLevel:
		color = "\033[36m" // Cyan
		prefix = "[*]"
	case logrus.WarnLevel:
		color = "\033[33m" // Yellow
		prefix = "[!]"
	case logrus.ErrorLevel:
		color = "\033[31m" // Red
		prefix = "[-]"
	case logrus.FatalLevel:
		color = "\033[31m" // Red
		prefix = "[-]"
	default:
		color = "\033[36m" // Cyan
		prefix = "[*]"
	}
	
	reset := "\033[0m"
	
	message := entry.Message
	if f.verbose && len(entry.Data) > 0 {
		message = fmt.Sprintf("%s %v", message, entry.Data)
	}
	
	return []byte(fmt.Sprintf("%s%s%s %s\n", color, prefix, reset, message)), nil
}

// Success logs a success message with green color
func Success(logger Logger, message string, args ...interface{}) {
	green := "\033[32m"
	reset := "\033[0m"
	formatted := fmt.Sprintf(message, args...)
	fmt.Printf("%s[+]%s %s\n", green, reset, formatted)
}

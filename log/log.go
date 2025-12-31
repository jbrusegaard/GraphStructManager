package log

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

func InitializeLogger() *log.Logger {
	// Initialize the logger with default settings
	// This function can be expanded to include more complex initialization logic if needed

	styles := log.DefaultStyles()
	styles.Keys["role"] = lipgloss.NewStyle().Foreground(lipgloss.Color("#f305f0")).Bold(true)
	styles.Values["role"] = lipgloss.NewStyle().Bold(true)

	logger := log.New(os.Stdout)
	logLevel := os.Getenv("GSM_LOG_LEVEL")
	switch logLevel {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	case "fatal":
		logger.SetLevel(log.FatalLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}
	logger.SetTimeFormat("2006-01-02 15:04:05")

	logger.SetStyles(styles)

	return logger
}

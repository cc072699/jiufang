// Package logger provides structured logging using zap.
// This package initializes the global logger for the application.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerConfig contains the configuration for the logger.
type LoggerConfig struct {
	// Level is the log level (debug, info, warn, error, fatal)
	Level string `yaml:"level"`

	// Encoding is the log encoding (json, console)
	Encoding string `yaml:"encoding"`

	// OutputPaths is the list of output paths (stdout, stderr, file paths)
	OutputPaths []string `yaml:"output_paths"`

	// ErrorOutputPaths is the list of error output paths
	ErrorOutputPaths []string `yaml:"error_output_paths"`
}

// DefaultLoggerConfig returns the default logger configuration.
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:            "info",
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// Initialize creates and returns a new zap logger based on the configuration.
func Initialize(config *LoggerConfig) (*zap.Logger, error) {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	// Parse log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder
	var encoder zapcore.Encoder
	if config.Encoding == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create write syncer
	var writeSyncers []zapcore.WriteSyncer
	for _, path := range config.OutputPaths {
		switch path {
		case "stdout":
			writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
		case "stderr":
			writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stderr))
		default:
			// File path
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			writeSyncers = append(writeSyncers, zapcore.AddSync(file))
		}
	}

	// Create error write syncer
	var errorWriteSyncers []zapcore.WriteSyncer
	for _, path := range config.ErrorOutputPaths {
		switch path {
		case "stdout":
			errorWriteSyncers = append(errorWriteSyncers, zapcore.AddSync(os.Stdout))
		case "stderr":
			errorWriteSyncers = append(errorWriteSyncers, zapcore.AddSync(os.Stderr))
		default:
			// File path
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			errorWriteSyncers = append(errorWriteSyncers, zapcore.AddSync(file))
		}
	}

	// Create core
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncers...), level),
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(errorWriteSyncers...), zapcore.ErrorLevel),
	)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger, nil
}

// InitializeDefault creates and returns a default zap logger.
func InitializeDefault() (*zap.Logger, error) {
	return Initialize(DefaultLoggerConfig())
}

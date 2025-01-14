package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
)

var logger *slog.Logger

// pipeToLogger writes data from a Reader to a logger with a specified level.
func pipeToLogger(reader io.Reader, logger *slog.Logger, level slog.Level) {
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			// Trim and log the output
			line := bytes.TrimSpace(buf[:n])
			logger.Log(context.Background(), level, string(line))
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Error("Error reading pipe", "error", err)
			break
		}
	}
}


// ConvertLogLevel converts a string log level to slog.HandlerOptions.
func ConvertLogLevel(levelStr string) (*slog.HandlerOptions, error) {
	var level slog.Level
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN", "WARNING":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		return nil, errors.New("invalid log level: " + levelStr)
	}

	return &slog.HandlerOptions{
		Level: level,
	}, nil
}

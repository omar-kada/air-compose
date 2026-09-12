package logs

import (
	"context"
	"io"
	"log/slog"
	"omar-kada/air-compose/internal/models"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/natefinch/lumberjack"
)

func newRotatingTempLogWriter(path string, maxSizeMB int) (io.Writer, error) {

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		slog.Warn("unable to create log file, logging to stdout only", "error", err)
		return os.Stdout, err
	}
	if err := file.Close(); err != nil {
		slog.Warn("unable to close log file handle", "path", path, "error", err)
	}

	logger := &lumberjack.Logger{
		Filename: path,
		MaxSize:  maxSizeMB,
		Compress: true,
	}

	slog.Info("logging to persistent file", "path", path, "max_size_mb", maxSizeMB)
	return logger, nil
}

// NewDefaultLogger creates a new slog.Logger with file rotation and broadcasting capabilities.
func NewDefaultLogger(logsHub Hub, logParams models.LoggerParams) *slog.Logger {

	logWriter, err := newRotatingTempLogWriter(logParams.LogFile, logParams.MaxLogFileSize)
	if err != nil {
		if existing, err := LoadRecordsFromFile(logParams.LogFile); err != nil {
			slog.Warn("unable to load existing log records", "path", logParams.LogFile, "error", err)
		} else if len(existing) > 0 {
			logsHub.Init(existing)
		}
	}
	var base slog.Handler
	if logParams.IsHumanLog() {
		base = slog.NewMultiHandler(
			tint.NewTextHandler(os.Stdout, &tint.Options{
				Level:      logParams.ToLevel(),
				TimeFormat: time.Kitchen,
				AddSource:  true,
			}),
			slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: logParams.ToLevel()}),
		)
	} else {
		base = slog.NewJSONHandler(io.MultiWriter(os.Stdout, logWriter), &slog.HandlerOptions{Level: logParams.ToLevel()})
	}

	tapLogHandler := NewTapHandler(base, func(_ context.Context, r slog.Record) {
		logsHub.Broadcast(r)
	})
	return slog.New(tapLogHandler)
}

// InitLogHub initializes a new HistoryHub with the given parameters
func InitLogHub(logParams models.LoggerParams) *HistoryHub {

	logHub := NewHistoryHub(logParams.MaxLogHistorySize)
	logs, err := LoadRecordsFromFile(logParams.LogFile)
	if err == nil {
		logHub.Init(logs)
	} else {
		slog.Warn("Couldn't load logs from logfile", "logpath", logParams.LogFile, "err", err)
	}
	slog.SetDefault(NewDefaultLogger(logHub, logParams))
	return logHub
}

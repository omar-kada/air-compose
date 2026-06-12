package socket

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"omar-kada/air-compose/api"
)

// LogHandler handles log-related messages and connection events
type LogHandler struct {
	logger *slog.Logger
	sender MessageSender

	mu       sync.Mutex
	stopLogs context.CancelFunc
}

// NewLogHandler creates a new LogHandler instance
func NewLogHandler(logger *slog.Logger, sender MessageSender) *LogHandler {
	return &LogHandler{
		logger: logger,
		sender: sender,
	}
}

// OnConnect is called when a new connection is established
func (lh *LogHandler) OnConnect(_ context.Context) {
	lh.logger.Debug("[SOCKET] log handler connected")
}

// HandleMessage is called when a new message is received
func (lh *LogHandler) HandleMessage(ctx context.Context, msg any) {

	switch m := msg.(type) {
	case api.ClientMessageStartLogs:
		lh.HandleStartLog(ctx, m.Value)
	case api.ClientMessageEndLogs:
		lh.CancelLogs()
	default:
		// Ignore messages that are not related to logs
	}
}

// HandleStartLog handles the start log message
func (lh *LogHandler) HandleStartLog(ctx context.Context, msg api.StartLogsMessage) {
	lh.logger.Info("[SOCKET] started streaming logs", "previousLines", msg.PreviousLines)
	lh.CancelLogs()

	logCtx, cancel := context.WithCancel(ctx)

	lh.mu.Lock()
	lh.stopLogs = cancel
	lh.mu.Unlock()

	go func() {
		defer lh.CancelLogs()
		for lines := range SubscribeToLogs(logCtx, int(msg.PreviousLines)) {
			if len(lines) > 1 {

				var messages []api.LogLine
				for _, line := range lines {
					messages = append(messages, api.LogLine{
						Msg:   line.Message,
						Level: line.Level.String(),
						Time:  line.Time,
					})
				}
				if err := lh.sender.SendPreviousLogs(logCtx, messages); err != nil {
					return
				}
			} else if len(lines) == 1 {
				if err := lh.sender.SendLog(logCtx, api.LogLine{
					Msg:   lines[0].Message,
					Level: lines[0].Level.String(),
					Time:  lines[0].Time,
				}); err != nil {
					return
				}
			}
		}
	}()
}

// SubscribeToLogs subscribes to log messages
func SubscribeToLogs(ctx context.Context, previousLines int) <-chan []slog.Record {
	ch := make(chan []slog.Record)
	go func() {
		defer close(ch)
		// Send previous lines
		if previousLines > 0 {
			records := make([]slog.Record, previousLines)
			for i := 0; i < previousLines; i++ {
				records[i] = slog.Record{
					Time:    time.Now().Add(-time.Duration(previousLines-i) * time.Second),
					Message: fmt.Sprintf("Previous log line %d", i+1),
					Level:   slog.LevelInfo,
				}
			}
			select {
			case ch <- records:
			case <-ctx.Done():
				return
			}
		}

		// Send new log lines every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		counter := 1
		for {
			select {
			case <-ticker.C:
				record := slog.Record{
					Time:    time.Now(),
					Message: fmt.Sprintf("New log line %d", counter),
					Level:   slog.LevelInfo,
				}
				select {
				case ch <- []slog.Record{record}:
					counter++
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

// CancelLogs cancels the active log stream
func (lh *LogHandler) CancelLogs() {
	lh.logger.Info("[SOCKET] cancelling active log stream")
	lh.mu.Lock()
	defer lh.mu.Unlock()
	if lh.stopLogs != nil {
		lh.stopLogs()
		lh.stopLogs = nil
	}
}

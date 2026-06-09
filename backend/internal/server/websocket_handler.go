package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"omar-kada/air-compose/api"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

var sessionIDCounter atomic.Uint64

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("[SOCKET] websocket upgrade failed", "error", err)
		return
	}

	sessionID := sessionIDCounter.Add(1)
	sess := &session{id: sessionID, conn: conn, logger: slog.With("session", sessionID)}
	slog.Info("[SOCKET] websocket upgrade succeeded", "session", sessionID, "remoteAddr", r.RemoteAddr)
	sess.run(r.Context())
}

// ─── session ──────────────────────────────────────────────────────────────────

type session struct {
	id       uint64
	conn     *websocket.Conn
	logger   *slog.Logger
	mu       sync.Mutex
	stopLogs context.CancelFunc
}

func (sess *session) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		sess.logger.Info("[SOCKET] session closing")
		cancel()
		sess.cancelLogs()
		sess.conn.Close(websocket.StatusNormalClosure, "")
	}()

	sess.logger.Debug("[SOCKET] session started")
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				state := api.StateEvent{
					Status: "running",
				}
				if err := wsjson.Write(ctx, sess.conn, state); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	sess.readLoop(ctx)
}

func (sess *session) readLoop(ctx context.Context) {
	for {
		var raw json.RawMessage
		if err := wsjson.Read(ctx, sess.conn, &raw); err != nil {
			sess.logger.Debug("[SOCKET] read stopping", "cause", err)
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure && websocket.CloseStatus(err) != websocket.StatusGoingAway {
				sess.logger.Error("[SOCKET] read error", "error", err)
			}
			return
		}
		sess.logger.Info("[SOCKET] message received", "msg", string(raw))

		var envelope struct {
			Kind  string          `json:"kind"`
			Value json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			sess.logger.Error("[SOCKET] envelope error", "error", err)
			return
		}

		sess.dispatch(ctx, envelope.Kind, envelope.Value)
	}
}

func (sess *session) dispatch(ctx context.Context, messageType string, raw json.RawMessage) {
	switch messageType {
	case string(api.ClientEventStartLogsKindStartLogs):
		var msg api.StartLogsEvent
		if err := json.Unmarshal(raw, &msg); err != nil {
			sess.logger.Error("[SOCKET] bad message structure", "type", messageType, "message", string(raw), "error", err)
			return
		}
		sess.handleStartLog(ctx, msg)

	case string(api.ClientEventEndLogsKindEndLogs):
		sess.cancelLogs()

	default:
		sess.logger.Warn("[SOCKET] unknown message type", "type", messageType)
	}
}

// ─── handlers ─────────────────────────────────────────────────────────────────

func (sess *session) handleStartLog(ctx context.Context, msg api.StartLogsEvent) {
	sess.logger.Info("[SOCKET] started streaming logs", "previousLines", msg.PreviousLines)
	sess.cancelLogs()

	logCtx, cancel := context.WithCancel(ctx)

	sess.mu.Lock()
	sess.stopLogs = cancel
	sess.mu.Unlock()

	go func() {
		defer cancel()
		for lines := range subscribeToLogs(logCtx, int(msg.PreviousLines)) {
			if len(lines) > 1 {

				var messages []api.LogLine
				for _, line := range lines {
					messages = append(messages, api.LogLine{
						Msg:   line.Message,
						Level: line.Level.String(),
						Time:  line.Time,
					})
				}
				if err := wsjson.Write(logCtx, sess.conn, api.ServerEventPreviousLogs{
					Kind:  api.ServerEventPreviousLogsKindPreviousLogs,
					Value: messages,
				}); err != nil {
					return
				}
			} else if len(lines) == 1 {
				if err := wsjson.Write(logCtx, sess.conn, api.ServerEventLog{
					Kind: api.ServerEventLogKindLog,
					Value: api.LogLine{
						Msg:   lines[0].Message,
						Level: lines[0].Level.String(),
						Time:  lines[0].Time,
					},
				}); err != nil {
					return
				}
			}
		}
	}()
}

func subscribeToLogs(ctx context.Context, previousLines int) <-chan []slog.Record {
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

func (sess *session) cancelLogs() {
	sess.logger.Info("[SOCKET] cancelling active log stream")
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.stopLogs != nil {
		sess.stopLogs()
		sess.stopLogs = nil
	}
}

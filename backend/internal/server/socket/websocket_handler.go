// Package socket provides WebSocket connection handling and message processing
package socket

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"omar-kada/air-compose/api"
	"sync/atomic"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

var sessionIDCounter atomic.Uint64

// WebSocketHandler upgrades the request to a websocket and starts a session.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("[SOCKET] websocket upgrade failed", "error", err)
		return
	}

	sessionID := sessionIDCounter.Add(1)
	sender := NewWebSocketMessageSender(conn)
	logger := slog.With("session", sessionID)
	sess := &session{
		id:       sessionID,
		conn:     conn,
		logger:   logger,
		sender:   sender,
		handlers: []Handler{NewLogHandler(logger, sender), NewStateHandler(logger, sender)},
	}
	slog.Info("[SOCKET] websocket upgrade succeeded", "session", sessionID, "remoteAddr", r.RemoteAddr)
	sess.run(r.Context())
}

// ─── session ──────────────────────────────────────────────────────────────────

type session struct {
	id       uint64
	conn     *websocket.Conn
	logger   *slog.Logger
	sender   MessageSender
	handlers []Handler
}

func (sess *session) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		sess.logger.Info("[SOCKET] session closing")
		cancel()
		sess.conn.Close(websocket.StatusNormalClosure, "session ended")
	}()

	// Call OnConnect for each handler
	for _, handler := range sess.handlers {
		handler.OnConnect(ctx)
	}

	sess.ReadLoop(ctx)
}

// ReadLoop reads messages from the websocket connection
func (sess *session) ReadLoop(ctx context.Context) {
	for {
		msg, err := sess.ReadMessage(ctx)
		if err != nil {
			return
		}

		if msg == nil {
			continue
		}

		// Call HandleMessage for each handler
		for _, handler := range sess.handlers {
			handler.HandleMessage(ctx, msg)
		}
	}
}

// ReadMessage reads and extracts the message type and payload
func (sess *session) ReadMessage(ctx context.Context) (any, error) {
	var raw api.ClientMessage
	if err := wsjson.Read(ctx, sess.conn, &raw); err != nil {
		sess.logger.Debug("[SOCKET] read stopping", "cause", err)
		if websocket.CloseStatus(err) != websocket.StatusNormalClosure && websocket.CloseStatus(err) != websocket.StatusGoingAway {
			sess.logger.Error("[SOCKET] read error", "error", err)
		}
		return nil, err
	}

	kind, err := raw.Discriminator()
	if err != nil || kind == "" {
		sess.logger.Error("[SOCKET] error reading message kind", "error", err)
		sess.sender.SendError(ctx, api.Error{
			Code:    api.ErrorCodeINVALIDREQUEST,
			Message: fmt.Sprintf("error reading message kind: %v", raw),
		})
		return nil, nil
	}

	value, err := raw.ValueByDiscriminator()
	if err != nil {
		sess.logger.Error("[SOCKET] error reading message", "error", err)
		sess.sender.SendError(ctx, api.Error{
			Code:    api.ErrorCodeINVALIDREQUEST,
			Message: fmt.Sprintf("error reading message: %v", raw),
		})
		return nil, nil
	}
	sess.logger.Info("[SOCKET] message received", "kind", kind, "value", value)

	return value, nil
}

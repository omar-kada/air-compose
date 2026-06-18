// Package socket provides WebSocket connection handling and message processing
package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"omar-kada/air-compose/api"
	"omar-kada/air-compose/internal/models"
	"sync"
	"sync/atomic"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// WebSocketHandler provides WebSocket connection handling and message processing
type WebSocketHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
	BroadcastEvent(ctx context.Context, event models.Event)
}

type websocketHandler struct {
	sessionIDCounter atomic.Uint64
	sessions         map[uint64]*session
	mu               sync.RWMutex
}

// NewWebSocketHandler creates a new WebSocketHandler instance.
func NewWebSocketHandler() WebSocketHandler {
	return &websocketHandler{
		sessions: make(map[uint64]*session),
	}
}

// Handle upgrades the request to a websocket and starts a session.
func (h *websocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("[SOCKET] websocket upgrade failed", "error", err)
		return
	}

	sessionID := h.sessionIDCounter.Add(1)
	sender := NewWebSocketMessageSender(conn)
	logger := slog.With("session", sessionID)
	ctx, cancel := context.WithCancel(r.Context())
	sess := &session{
		id:       sessionID,
		conn:     conn,
		logger:   logger,
		sender:   sender,
		handlers: []Handler{NewLogHandler(logger, sender)},
		ctx:      ctx,
		cancel:   cancel,
	}
	slog.Debug("[SOCKET] websocket upgrade succeeded", "session", sessionID, "remoteAddr", r.RemoteAddr)

	h.mu.Lock()
	h.sessions[sessionID] = sess
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.sessions, sessionID)
		h.mu.Unlock()
	}()

	sess.run()
}

func (h *websocketHandler) BroadcastEvent(_ context.Context, event models.Event) {
	// Serialize the event to JSON
	eventJSON, err := json.Marshal(api.ServerMessageEvent{
		Kind: api.ServerMessageEventKindEvent,
		Value: api.EventMessage{
			Msg:            event.Msg,
			Type:           api.EventType(event.Type),
			DeploymentId:   &event.ObjectID,
			IsNotification: event.IsNotification,
		}})
	if err != nil {
		slog.Error("[SOCKET] failed to serialize event", "error", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, session := range h.sessions {
		go session.SendRawText(eventJSON)
	}
}

// ─── session ──────────────────────────────────────────────────────────────────

type session struct {
	id       uint64
	conn     *websocket.Conn
	logger   *slog.Logger
	sender   MessageSender
	handlers []Handler
	ctx      context.Context
	cancel   context.CancelFunc
}

func (sess *session) run() {

	defer func() {
		sess.logger.Debug("[SOCKET] session closing")
		sess.cancel()
		sess.conn.Close(websocket.StatusNormalClosure, "session ended")
	}()

	// Call OnConnect for each handler
	for _, handler := range sess.handlers {
		handler.OnConnect(sess.ctx)
	}

	sess.ReadLoop(sess.ctx)
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

func (sess *session) SendRawText(data []byte) {
	sess.conn.Write(sess.ctx, websocket.MessageText, data)
}

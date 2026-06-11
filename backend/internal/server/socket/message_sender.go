package socket

import (
	"context"
	"omar-kada/air-compose/api"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// MessageSender defines an interface for sending messages
type MessageSender interface {
	SendStateMessage(ctx context.Context, state api.StateMessage) error
	SendPreviousLogs(ctx context.Context, logs api.LogMessages) error
	SendLog(ctx context.Context, log api.LogLine) error
	SendError(ctx context.Context, err api.Error) error
}

// WebSocketMessageSender implements MessageSender for WebSocket
type WebSocketMessageSender struct {
	conn *websocket.Conn
}

// NewWebSocketMessageSender creates a new WebSocketMessageSender instance
func NewWebSocketMessageSender(conn *websocket.Conn) *WebSocketMessageSender {
	return &WebSocketMessageSender{conn: conn}
}

// SendStateMessage sends a state message
func (ws *WebSocketMessageSender) SendStateMessage(ctx context.Context, state api.StateMessage) error {
	return wsjson.Write(ctx, ws.conn, api.ServerMessageState{
		Kind:  api.ServerMessageStateKindState,
		Value: state,
	})
}

// SendPreviousLogs sends previous logs
func (ws *WebSocketMessageSender) SendPreviousLogs(ctx context.Context, logs api.LogMessages) error {
	return wsjson.Write(ctx, ws.conn, api.ServerMessagePreviousLogs{
		Kind:  api.ServerMessagePreviousLogsKindPreviousLogs,
		Value: logs,
	})
}

// SendLog sends a log message
func (ws *WebSocketMessageSender) SendLog(ctx context.Context, log api.LogLine) error {
	return wsjson.Write(ctx, ws.conn, api.ServerMessageLog{
		Kind:  api.ServerMessageLogKindLog,
		Value: log,
	})
}

// SendError sends an error message
func (ws *WebSocketMessageSender) SendError(ctx context.Context, err api.Error) error {
	return wsjson.Write(ctx, ws.conn, api.ServerMessageError{
		Kind:  api.ServerMessageErrorKindError,
		Value: err,
	})
}

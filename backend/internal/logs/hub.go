// Package logs provides functionality for broadcasting log messages to connected clients
package logs

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Client represents one connected websocket subscriber.
type Client struct {
	Send chan []slog.Record
}

// Hub fans out log lines to all connected clients
type Hub interface {
	Broadcast(msg slog.Record)
	Register() *Client
	Unregister(c *Client)
	Init([]slog.Record)
}

const clientSendTimeout = 10 * time.Millisecond

// HistoryHub fans out log lines to all connected clients and keeps a fixed-size
// ring buffer of recent lines for newly connecting clients.
type HistoryHub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}

	historyMu sync.Mutex
	history   []slog.Record // fixed size = maxHistory, ring buffer
	next      int           // index the next write lands on
	count     int           // valid entries so far, caps at len(history)
}

// NewHistoryHub creates a HistoryHub that retains up to maxHistory
// Pass 0 to disable history (new clients only see lines from the moment they connect).
func NewHistoryHub(maxHistory int) *HistoryHub {
	return &HistoryHub{
		clients: make(map[*Client]struct{}),
		history: make([]slog.Record, maxHistory),
	}
}

// Broadcast sends msg to every connected client and records it in history.
// msg must not be mutated by the caller after this call (ownership transfers
// to the Hub / its clients / its history buffer).
func (h *HistoryHub) Broadcast(msg slog.Record) {
	h.addToHistory(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.Send <- []slog.Record{msg}:
		case <-time.After(clientSendTimeout):
			slog.NewTextHandler(os.Stdout, nil).Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelWarn, "slow log client detected", 0))
			// Slow client — drop rather than block the logger or other clients.
		}
	}
}

// Register adds a new client, replays buffered history to it, and returns it.
// Callers must call Unregister when done (typically via defer).
func (h *HistoryHub) Register() *Client {
	c := &Client{Send: make(chan []slog.Record, 256)}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	select {
	case c.Send <- h.snapshotHistory():
	default:
		slog.NewTextHandler(os.Stdout, nil).Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelWarn, "channel buffer full", 0))
		// Buffer full even before the client did anything — extremely
		// unlikely (256 cap), but don't block Register on it.
	}
	return c
}

// Unregister removes a client and closes its send channel.
func (h *HistoryHub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.Send)
	}
}

// Init initializes the history buffer with the given records, replacing any existing history.
func (h *HistoryHub) Init(records []slog.Record) {
	if len(h.history) == 0 {
		return
	}

	h.historyMu.Lock()
	defer h.historyMu.Unlock()

	h.next = 0
	h.count = 0

	if len(records) > len(h.history) {
		records = records[len(records)-len(h.history):]
	}
	if len(records) == 0 {
		return
	}

	copy(h.history, records)
	h.count = len(records)
	if h.count < len(h.history) {
		h.next = h.count
		return
	}
}

func (h *HistoryHub) addToHistory(msg slog.Record) {
	if len(h.history) == 0 {
		return
	}
	h.historyMu.Lock()
	h.history[h.next] = msg
	h.next = (h.next + 1) % len(h.history)
	if h.count < len(h.history) {
		h.count++
	}
	h.historyMu.Unlock()
}

// snapshotHistory returns buffered messages in chronological order.
func (h *HistoryHub) snapshotHistory() []slog.Record {
	h.historyMu.Lock()
	defer h.historyMu.Unlock()

	out := make([]slog.Record, h.count)
	if h.count < len(h.history) {
		copy(out, h.history[:h.count]) // not full yet, already in order
		return out
	}
	n := copy(out, h.history[h.next:]) // oldest..end
	copy(out[n:], h.history[:h.next])  // start..newest
	return out
}

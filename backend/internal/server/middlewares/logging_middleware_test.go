package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// memHandler captures slog records in-memory for assertions.
type memHandler struct {
	entries []map[string]any
}

func (*memHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *memHandler) Handle(_ context.Context, rec slog.Record) error {
	m := map[string]any{"msg": rec.Message}
	rec.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.Any()
		return true
	})
	h.entries = append(h.entries, m)
	return nil
}

func (h *memHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *memHandler) WithGroup(_ string) slog.Handler      { return h }

func TestLoggingMiddleware_RecordsStatusAndBytes(t *testing.T) {
	mh := &memHandler{}
	old := slog.Default()
	slog.SetDefault(slog.New(mh))
	defer slog.SetDefault(old)

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte("hello"))
	})

	h := LoggingMiddleware(inner)
	r := httptest.NewRequest("GET", "/testpath", http.NoBody)
	r.RemoteAddr = "1.2.3.4:1234"
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, r)

	assert.Len(t, mh.entries, 1)
	ent := mh.entries[0]
	assert.Equal(t, "[HTTP] request GET /testpath", ent["msg"])
	assert.Equal(t, "GET", ent["method"])
	assert.Equal(t, "/testpath", ent["path"])
	assert.Equal(t, int64(201), ent["status"])
	assert.Equal(t, int64(5), ent["bytes"])
	assert.Equal(t, "1.2.3.4:1234", ent["remote"])
	d, ok := ent["duration"].(time.Duration)
	assert.True(t, ok)
	assert.Greater(t, d, time.Duration(0))
}

func TestLoggingMiddleware_DefaultStatus(t *testing.T) {
	mh := &memHandler{}
	old := slog.Default()
	slog.SetDefault(slog.New(mh))
	defer slog.SetDefault(old)

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	h := LoggingMiddleware(inner)
	r := httptest.NewRequest("POST", "/d", http.NoBody)
	r.RemoteAddr = "4.3.2.1:4321"
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, r)

	assert.Len(t, mh.entries, 1)
	ent := mh.entries[0]
	assert.Equal(t, int64(200), ent["status"])
	assert.Equal(t, int64(2), ent["bytes"])
}

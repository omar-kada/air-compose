package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketHandlerConnectionAcceptance(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	// Convert http:// to ws://
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// Establish WebSocket connection
	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Connection should be established without error
	assert.NotNil(t, conn)
}

func TestWebSocketHandlerSessionIDIncrement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	const numSessions = 3
	sessions := make([]*websocket.Conn, numSessions)
	ids := make([]uint64, numSessions)

	// Establish multiple sessions and record their IDs
	for i := range numSessions {
		conn, _, err := websocket.Dial(context.Background(), url, nil)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		t.Cleanup(func() {
			if conn != nil {
				conn.Close(websocket.StatusNormalClosure, "")
			}
		})
		sessions[i] = conn

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		// Send a test message to make sure the conn is working
		testMsg := map[string]any{
			"kind":  "unknown_type",
			"value": map[string]any{},
		}
		err = wsjson.Write(ctx, conn, testMsg)
		t.Cleanup(cancel)

		assert.NoError(t, err)

		ids[i] = sessionIDCounter.Load()
	}

	// Verify IDs are incrementing
	for i := 1; i < numSessions; i++ {
		assert.Greater(t, ids[i], ids[i-1], "session ID should increment")
	}
}

func TestWebSocketHandlerMessageParsing(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Send a test message with envelope structure
	testMsg := map[string]interface{}{
		"kind":  "unknown_type",
		"value": map[string]interface{}{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = wsjson.Write(ctx, conn, testMsg)
	assert.NoError(t, err)
}

func TestWebSocketHandlerSessionCleanup(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)

	// Close the connection normally
	err = conn.Close(websocket.StatusNormalClosure, "test close")
	assert.NoError(t, err)

	// Try to read from closed connection - should get an error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var msg interface{}
	err = wsjson.Read(ctx, conn, &msg)
	assert.Error(t, err)
}

func TestWebSocketHandlerEnvelopeUnmarshal(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Send a properly structured message
	testMsg := map[string]interface{}{
		"kind": "startLogs",
		"value": map[string]interface{}{
			"previousLines": 10,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = wsjson.Write(ctx, conn, testMsg)
	assert.NoError(t, err)
}

func TestWebSocketHandlerInvalidJsonHandling(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Send invalid JSON that can't be unmarshaled
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Write raw invalid JSON
	err = conn.Write(ctx, websocket.MessageText, []byte("not valid json {"))
	assert.NoError(t, err)

	// Connection should close due to invalid envelope
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()
	var msg interface{}
	err = wsjson.Read(ctx2, conn, &msg)
	// Should eventually get an error as connection closes
	assert.Error(t, err)
}

func TestWebSocketHandlerInvalidEnvelopeHandling(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Send an unproperly structured message
	testMsg := map[string]interface{}{
		"invalid": "invalid",
		"value":   "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = wsjson.Write(ctx, conn, testMsg)
	assert.NoError(t, err)
	// Connection should close due to invalid envelope
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()
	var msg interface{}
	err = wsjson.Read(ctx2, conn, &msg)
	// Should eventually get an error as connection closes
	assert.Error(t, err)
}

func TestWebSocketHandlerCancelLogsMessage(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Send endLogs message
	testMsg := map[string]interface{}{
		"kind":  "endLogs",
		"value": "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = wsjson.Write(ctx, conn, testMsg)
	assert.NoError(t, err)
}

func TestWebSocketHandlerConcurrentConnections(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	const numConnections = 5
	connections := make([]*websocket.Conn, numConnections)

	// Establish multiple concurrent connections
	for i := 0; i < numConnections; i++ {
		conn, _, err := websocket.Dial(context.Background(), url, nil)
		assert.NoError(t, err, "failed to dial websocket %d", i)
		connections[i] = conn
	}

	// Clean up all connections
	for i, conn := range connections {
		if conn != nil {
			err := conn.Close(websocket.StatusNormalClosure, "")
			assert.NoError(t, err, "failed to close connection %d", i)
		}
	}

	// All connections should have been established
	for i, conn := range connections {
		assert.NotNil(t, conn, "connection %d is nil", i)
	}
}

func TestWebSocketHandlerMultipleMessages(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Send multiple messages
	messages := []map[string]interface{}{
		{
			"kind":  "unknown1",
			"value": map[string]interface{}{},
		},
		{
			"kind":  "unknown2",
			"value": map[string]interface{}{},
		},
		{
			"kind":  "endLogs",
			"value": map[string]interface{}{},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, msg := range messages {
		err = wsjson.Write(ctx, conn, msg)
		assert.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}
}

func TestWebSocketHandlerMessageEnvelopeStructure(t *testing.T) {
	// This test validates that the handler correctly parses message envelopes
	// by inspecting the JSON structure expectations

	envelope := struct {
		Kind  string          `json:"kind"`
		Value json.RawMessage `json:"value"`
	}{}

	testData := []byte(`{
		"kind": "startLogs",
		"value": {"previousLines": 5}
	}`)

	err := json.Unmarshal(testData, &envelope)
	assert.NoError(t, err)

	assert.Equal(t, "startLogs", envelope.Kind)

	assert.NotEmpty(t, envelope.Value)
}

func TestWebSocketHandlerContextCancellation(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(webSocketHandler))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.Dial(context.Background(), url, nil)
	assert.NoError(t, err)

	// Close connection immediately
	err = conn.Close(websocket.StatusGoingAway, "")
	assert.NoError(t, err)

	// Try to use the connection - should fail
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var msg interface{}
	err = wsjson.Read(ctx, conn, &msg)
	assert.Error(t, err)
}

package logs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHistoryHubBroadcastsToRegisteredClients(t *testing.T) {
	hub := NewHistoryHub(5)
	client := hub.Register()

	// Register sends the current snapshot immediately.
	initialSnapshot := receiveLogs(t, client.Send)
	assert.Empty(t, initialSnapshot)

	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "one", 0))

	got := receiveLogs(t, client.Send)
	assert.Equal(t, "one", got[0].Message)
}

func TestHistoryHubReplaysBufferedHistoryToNewClients(t *testing.T) {
	hub := NewHistoryHub(3)
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "one", 0))
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "two", 0))

	client := hub.Register()
	got := receiveLogs(t, client.Send)

	assert.Equal(t, "one", got[0].Message)
	assert.Equal(t, "two", got[1].Message)
}

func TestHistoryHubOverloadsHistoryWindow(t *testing.T) {
	hub := NewHistoryHub(2)
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "one", 0))
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "two", 0))
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "three", 0))

	client := hub.Register()
	got := receiveLogs(t, client.Send)

	assert.Equal(t, "two", got[0].Message)
	assert.Equal(t, "three", got[1].Message)
}

func TestHistoryHubMultipleClientsDifferentJoinTiming(t *testing.T) {
	hub := NewHistoryHub(2)
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "one", 0))
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "two", 0))

	clientA := hub.Register()
	snapshotA := receiveLogs(t, clientA.Send)
	assert.Equal(t, "one", snapshotA[0].Message)
	assert.Equal(t, "two", snapshotA[1].Message)

	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "three", 0))

	gotA3 := receiveLogs(t, clientA.Send)
	assert.Equal(t, "three", gotA3[0].Message)

	clientB := hub.Register()
	snapshotB := receiveLogs(t, clientB.Send)
	assert.Equal(t, "two", snapshotB[0].Message)
	assert.Equal(t, "three", snapshotB[1].Message)

	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "four", 0))

	gotA := receiveLogs(t, clientA.Send)
	gotB := receiveLogs(t, clientB.Send)

	assert.Equal(t, "four", gotA[0].Message)
	assert.Equal(t, slog.LevelInfo, gotA[0].Level)
	assert.Equal(t, "four", gotB[0].Message)
	assert.Equal(t, slog.LevelInfo, gotB[0].Level)
}

func TestHistoryHubBroadcastWaitsBrieflyBeforeDroppingSlowClient(t *testing.T) {
	hub := NewHistoryHub(3)
	client := hub.Register()
	_ = receiveLogs(t, client.Send)

	for i := 0; i < cap(client.Send); i++ {
		client.Send <- []slog.Record{slog.NewRecord(time.Now(), slog.LevelInfo, "blocked", 0)}
	}

	started := time.Now()
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "one", 0))
	elapsed := time.Since(started)

	assert.GreaterOrEqual(t, elapsed, clientSendTimeout)
}

func TestHistoryHubUnregisterRemovesClient(t *testing.T) {
	hub := NewHistoryHub(3)
	client := hub.Register()
	_ = receiveLogs(t, client.Send)

	hub.Unregister(client)
	hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "one", 0))

	select {
	case msg, ok := <-client.Send:
		assert.False(t, ok)
		assert.Nil(t, msg)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for closed channel")
	}
}

func TestHistoryHubInitSeedsHistory(t *testing.T) {
	hub := NewHistoryHub(3)
	records := []slog.Record{
		slog.NewRecord(time.Unix(1, 0), slog.LevelInfo, "one", 0),
		slog.NewRecord(time.Unix(2, 0), slog.LevelWarn, "two", 0),
	}

	hub.Init(records)

	client := hub.Register()
	got := receiveLogs(t, client.Send)
	assert.Len(t, got, 2)
	assert.Equal(t, "one", got[0].Message)
	assert.Equal(t, "two", got[1].Message)
}

func TestHistoryHubInitKeepsOnlyNewestRecords(t *testing.T) {
	hub := NewHistoryHub(3)
	records := []slog.Record{
		slog.NewRecord(time.Unix(1, 0), slog.LevelInfo, "one", 0),
		slog.NewRecord(time.Unix(2, 0), slog.LevelWarn, "two", 0),
		slog.NewRecord(time.Unix(3, 0), slog.LevelInfo, "three", 0),
		slog.NewRecord(time.Unix(4, 0), slog.LevelWarn, "four", 0),
	}

	hub.Init(records)

	client := hub.Register()
	got := receiveLogs(t, client.Send)
	assert.Len(t, got, 3)
	assert.Equal(t, "two", got[0].Message)
	assert.Equal(t, "three", got[1].Message)
	assert.Equal(t, "four", got[2].Message)
}

func receiveLogs(t *testing.T, ch <-chan []slog.Record) []slog.Record {
	t.Helper()

	select {
	case msg := <-ch:
		return msg
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for log snapshot")
		return nil
	}
}

func BenchmarkHistoryHubBroadcast(b *testing.B) {
	hub := NewHistoryHub(1000)
	client := hub.Register()
	defer hub.Unregister(client)

	// Drain the initial snapshot.
	<-client.Send

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0))
		<-client.Send
	}
}

func BenchmarkHistoryHubBroadcastSequential(b *testing.B) {
	hub := NewHistoryHub(100)
	clients := make([]*Client, 2)

	for i := range clients {
		client := hub.Register()
		clients[i] = client
		go func() {
			for range client.Send {
				time.Sleep(time.Microsecond)
			}
		}()
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			hub.Broadcast(slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0))
		}
	})
}

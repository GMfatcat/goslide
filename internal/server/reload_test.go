package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

func TestBroadcaster_SendReload(t *testing.T) {
	b := newBroadcaster()
	srv := httptest.NewServer(http.HandlerFunc(b.handleWS))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	time.Sleep(50 * time.Millisecond)
	b.send(reloadMsg{Type: "reload"})

	_, msg, err := conn.Read(ctx)
	require.NoError(t, err)
	require.Contains(t, string(msg), `"type":"reload"`)
}

func TestBroadcaster_SendError(t *testing.T) {
	b := newBroadcaster()
	srv := httptest.NewServer(http.HandlerFunc(b.handleWS))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer conn.CloseNow()

	time.Sleep(50 * time.Millisecond)
	b.send(reloadMsg{Type: "error", Message: "parse failed"})

	_, msg, err := conn.Read(ctx)
	require.NoError(t, err)
	require.Contains(t, string(msg), `"type":"error"`)
	require.Contains(t, string(msg), "parse failed")
}

func TestBroadcaster_MultipleClients(t *testing.T) {
	b := newBroadcaster()
	srv := httptest.NewServer(http.HandlerFunc(b.handleWS))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer c1.CloseNow()

	c2, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)
	defer c2.CloseNow()

	time.Sleep(50 * time.Millisecond)

	b.send(reloadMsg{Type: "reload"})

	_, msg1, err := c1.Read(ctx)
	require.NoError(t, err)
	require.Contains(t, string(msg1), "reload")

	_, msg2, err := c2.Read(ctx)
	require.NoError(t, err)
	require.Contains(t, string(msg2), "reload")
}

func TestBroadcaster_ClientDisconnect(t *testing.T) {
	b := newBroadcaster()
	srv := httptest.NewServer(http.HandlerFunc(b.handleWS))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	conn.Close(websocket.StatusNormalClosure, "bye")
	time.Sleep(50 * time.Millisecond)

	b.send(reloadMsg{Type: "reload"})
}

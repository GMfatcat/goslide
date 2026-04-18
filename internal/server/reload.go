package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type reloadMsg struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

type broadcaster struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newBroadcaster() *broadcaster {
	return &broadcaster{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (b *broadcaster) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	b.mu.Lock()
	b.clients[conn] = struct{}{}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, conn)
		b.mu.Unlock()
		conn.CloseNow()
	}()

	for {
		_, msg, err := conn.Read(r.Context())
		if err != nil {
			return
		}
		var peek struct{ Type string `json:"type"` }
		if json.Unmarshal(msg, &peek) == nil && peek.Type == "presenter-slide" {
			b.broadcast(msg, conn)
		}
	}
}

func (b *broadcaster) broadcast(data []byte, sender *websocket.Conn) {
	b.mu.Lock()
	clients := make([]*websocket.Conn, 0, len(b.clients))
	for c := range b.clients {
		if c != sender {
			clients = append(clients, c)
		}
	}
	b.mu.Unlock()

	for _, c := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		c.Write(ctx, websocket.MessageText, data)
		cancel()
	}
}

func (b *broadcaster) send(msg reloadMsg) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	b.mu.Lock()
	clients := make([]*websocket.Conn, 0, len(b.clients))
	for c := range b.clients {
		clients = append(clients, c)
	}
	b.mu.Unlock()

	for _, c := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := c.Write(ctx, websocket.MessageText, data)
		cancel()
		if err != nil {
			log.Printf("[ws] write failed, removing client: %v", err)
			b.mu.Lock()
			delete(b.clients, c)
			b.mu.Unlock()
			c.CloseNow()
		}
	}
}

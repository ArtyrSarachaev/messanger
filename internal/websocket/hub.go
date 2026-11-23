package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeTimeout = 10 * time.Second

// client — одно WebSocket-соединение. У пользователя может быть
// несколько соединений (например, несколько вкладок).
// gorilla/websocket не допускает конкурентную запись, поэтому запись под мьютексом.
type client struct {
	userID string
	conn   *websocket.Conn
	mu     sync.Mutex
}

func (c *client) writeJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return c.conn.WriteJSON(v)
}

func (c *client) writeControl(messageType int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.WriteControl(messageType, nil, time.Now().Add(writeTimeout))
}

// hub хранит активные соединения по ID пользователя.
type hub struct {
	mu      sync.RWMutex
	clients map[string]map[*client]struct{}
}

func newHub() *hub {
	return &hub{clients: make(map[string]map[*client]struct{})}
}

func (h *hub) add(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[c.userID] == nil {
		h.clients[c.userID] = make(map[*client]struct{})
	}
	h.clients[c.userID][c] = struct{}{}
}

func (h *hub) remove(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.clients[c.userID]
	delete(conns, c)
	if len(conns) == 0 {
		delete(h.clients, c.userID)
	}
}

// connections возвращает копию списка соединений пользователя,
// чтобы запись шла без удержания блокировки хаба.
func (h *hub) connections(userID string) []*client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*client, 0, len(h.clients[userID]))
	for c := range h.clients[userID] {
		result = append(result, c)
	}
	return result
}

func (h *hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, conns := range h.clients {
		for c := range conns {
			_ = c.conn.Close()
		}
	}
}

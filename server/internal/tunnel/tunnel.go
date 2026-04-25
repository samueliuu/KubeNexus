package tunnel

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || origin == "http://localhost:3000" || origin == "http://localhost:3001"
	},
}

type MessageType string

const (
	TypeHeartbeat      MessageType = "heartbeat"
	TypeTask           MessageType = "task"
	TypeTunnelRequest  MessageType = "tunnel_request"
	TypeTunnelResponse MessageType = "tunnel_response"
)

type Message struct {
	Type    MessageType     `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

type TunnelPayload struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type TunnelResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type ClusterConnection struct {
	ClusterID    string
	Conn         *websocket.Conn
	LastPing     time.Time
	pendingReqs  map[string]chan *TunnelResponse
	mu           sync.Mutex
}

func (cc *ClusterConnection) SendTask(payload interface{}) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	msg := Message{
		Type:    TypeTask,
		ID:      uuid.New().String(),
		Payload: mustMarshal(payload),
	}
	return cc.Conn.WriteJSON(msg)
}

func (cc *ClusterConnection) SendTunnelRequest(id string, payload *TunnelPayload) (*TunnelResponse, error) {
	cc.mu.Lock()
	ch := make(chan *TunnelResponse, 1)
	cc.pendingReqs[id] = ch
	msg := Message{
		Type:    TypeTunnelRequest,
		ID:      id,
		Payload: mustMarshal(payload),
	}
	if err := cc.Conn.WriteJSON(msg); err != nil {
		delete(cc.pendingReqs, id)
		cc.mu.Unlock()
		return nil, err
	}
	cc.mu.Unlock()

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(30 * time.Second):
		cc.mu.Lock()
		delete(cc.pendingReqs, id)
		cc.mu.Unlock()
		return nil, http.ErrHandlerTimeout
	}
}

func (cc *ClusterConnection) handleMessage(msg *Message) {
	if msg.Type == TypeTunnelResponse {
		cc.mu.Lock()
		if ch, ok := cc.pendingReqs[msg.ID]; ok {
			var resp TunnelResponse
			if err := json.Unmarshal(msg.Payload, &resp); err != nil {
				log.Printf("Failed to unmarshal tunnel response: %v", err)
				ch <- &TunnelResponse{StatusCode: 500, Body: "internal unmarshal error"}
			} else {
				ch <- &resp
			}
			delete(cc.pendingReqs, msg.ID)
		}
		cc.mu.Unlock()
	}
}

type Manager struct {
	connections map[string]*ClusterConnection
	mu          sync.RWMutex
	onConnect   func(clusterID string)
	onDisconnect func(clusterID string)
}

func NewManager(onConnect, onDisconnect func(clusterID string)) *Manager {
	return &Manager{
		connections:  make(map[string]*ClusterConnection),
		onConnect:    onConnect,
		onDisconnect: onDisconnect,
	}
}

func (m *Manager) HandleWebSocket(c *gin.Context) {
	clusterID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed for cluster %s: %v", clusterID, err)
		return
	}

	cc := &ClusterConnection{
		ClusterID:   clusterID,
		Conn:        conn,
		LastPing:    time.Now(),
		pendingReqs: make(map[string]chan *TunnelResponse),
	}

	m.mu.Lock()
	if old, ok := m.connections[clusterID]; ok {
		old.Conn.Close()
	}
	m.connections[clusterID] = cc
	m.mu.Unlock()

	log.Printf("Cluster %s WebSocket connected", clusterID)
	if m.onConnect != nil {
		m.onConnect(clusterID)
	}

	defer func() {
		m.mu.Lock()
		if m.connections[clusterID] == cc {
			delete(m.connections, clusterID)
		}
		m.mu.Unlock()
		conn.Close()
		log.Printf("Cluster %s WebSocket disconnected", clusterID)
		if m.onDisconnect != nil {
			m.onDisconnect(clusterID)
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		cc.LastPing = time.Now()

		if msg.Type == TypeHeartbeat {
			continue
		}

		cc.handleMessage(&msg)
	}
}

func (m *Manager) IsConnected(clusterID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.connections[clusterID]
	return ok
}

func (m *Manager) SendTask(clusterID string, payload interface{}) error {
	m.mu.RLock()
	cc, ok := m.connections[clusterID]
	m.mu.RUnlock()
	if !ok {
		return ErrClusterNotConnected
	}
	return cc.SendTask(payload)
}

func (m *Manager) ProxyRequest(clusterID string, req *TunnelPayload) (*TunnelResponse, error) {
	m.mu.RLock()
	cc, ok := m.connections[clusterID]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrClusterNotConnected
	}
	id := uuid.New().String()
	return cc.SendTunnelRequest(id, req)
}

func (m *Manager) GetConnectedClusters() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ids []string
	for id := range m.connections {
		ids = append(ids, id)
	}
	return ids
}

var ErrClusterNotConnected = &ClusterNotConnectedError{}

type ClusterNotConnectedError struct{}

func (e *ClusterNotConnectedError) Error() string {
	return "cluster not connected via WebSocket"
}

func mustMarshal(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshal: %v", err)
		return json.RawMessage("{}")
	}
	return b
}

package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"maplocationshare/backend/models"
	"maplocationshare/backend/storage"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	HandshakeTimeout: 10 * time.Second,
}

type Client struct {
	SessionID string
	UserID    string
	Conn      *websocket.Conn
	Send      chan []byte
	mu         sync.Mutex
}

type Hub struct {
	clients    map[string][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

type Message struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	SessionID string          `json:"session_id,omitempty"`
}

var (
	hub   *Hub
	store storage.Storage
)

func InitWebSocketHub(r storage.Storage) {
	store = r
	hub = &Hub{
		clients:    make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
	go hub.run()
}

func HandleWebSocket(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		log.Printf("WebSocket连接失败: Session ID为空")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	log.Printf("收到WebSocket连接请求: SessionID=%s, URL=%s", sessionID, c.Request.URL.Path)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := &Client{
		SessionID: sessionID,
		UserID:    "",
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.SessionID] = append(h.clients[client.SessionID], client)
			h.mu.Unlock()
			log.Printf("客户端已注册: SessionID=%s, UserID=%s", client.SessionID, client.UserID)

			h.mu.RLock()
			clients := h.clients[client.SessionID]
			locations, err := store.GetAllLocations(client.SessionID)
			if err != nil {
				log.Printf("获取位置失败: %v", err)
			} else {
				allLocationsMsg := models.AllLocations{
					Type:      "all_locations",
					Data:      locations,
					SessionID: client.SessionID,
				}
				msgBytes, _ := json.Marshal(allLocationsMsg)
				for _, c := range clients {
					select {
					case c.Send <- msgBytes:
					default:
						close(c.Send)
					}
				}
			}
			h.mu.RUnlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.SessionID]; ok {
				for i, c := range clients {
					if c == client {
						h.clients[client.SessionID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(h.clients[client.SessionID]) == 0 {
					delete(h.clients, client.SessionID)
				}
			}
			h.mu.Unlock()
			close(client.Send)
			log.Printf("客户端已注销: SessionID=%s, UserID=%s", client.SessionID, client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, clients := range h.clients {
				for _, client := range clients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		log.Printf("收到Pong: SessionID=%s", c.SessionID)
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: SessionID=%s, Error=%v", c.SessionID, err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("解析消息失败: %v", err)
			continue
		}

		log.Printf("收到消息: SessionID=%s, Type=%s", c.SessionID, msg.Type)

		if msg.Type == "location_update" {
			var location models.Location
			if err := json.Unmarshal(msg.Data, &location); err != nil {
				log.Printf("解析位置数据失败: %v", err)
				continue
			}

			c.UserID = location.UserID
			log.Printf("更新用户ID: SessionID=%s, UserID=%s", c.SessionID, c.UserID)

			if err := store.UpdateLocation(c.SessionID, location); err != nil {
				log.Printf("更新位置失败: %v", err)
			}

			hub.broadcastToSession(c.SessionID, msg.Data)
		} else if msg.Type == "ping" {
			pongMsg, _ := json.Marshal(map[string]interface{}{
				"type":      "pong",
				"timestamp": time.Now().UnixMilli(),
			})
			c.Conn.WriteMessage(websocket.TextMessage, pongMsg)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("发送消息失败: SessionID=%s, Error=%v", c.SessionID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			pingMsg, _ := json.Marshal(map[string]interface{}{
				"type":      "ping",
				"timestamp": time.Now().UnixMilli(),
			})
			if err := c.Conn.WriteMessage(websocket.PingMessage, pingMsg); err != nil {
				log.Printf("发送Ping失败: SessionID=%s, Error=%v", c.SessionID, err)
				return
			}
		}
	}
}

func (h *Hub) broadcastToSession(sessionID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[sessionID]; ok {
		locations, err := store.GetAllLocations(sessionID)
		if err != nil {
			log.Printf("获取所有位置失败: %v", err)
			return
		}

		allLocationsMsg := models.AllLocations{
			Type:      "all_locations",
			Data:      locations,
			SessionID: sessionID,
		}
		msgBytes, err := json.Marshal(allLocationsMsg)
		if err != nil {
			log.Printf("序列化消息失败: %v", err)
			return
		}
		for _, client := range clients {
			select {
			case client.Send <- msgBytes:
			default:
				close(client.Send)
			}
		}
	}
}

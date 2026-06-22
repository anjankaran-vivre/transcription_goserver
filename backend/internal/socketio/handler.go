package socketio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"

	"transcription-goserver/internal/logging"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

var (
	clients   = make(map[*client]bool)
	clientsMu sync.RWMutex
)

func HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("[WS] Upgrade error: %v\n", err)
		return
	}

	cl := &client{conn: conn, send: make(chan []byte, 256)}
	clientsMu.Lock()
	clients[cl] = true
	clientsMu.Unlock()

	fmt.Printf("[WS] Client connected (%d total)\n", len(clients))
	logging.GetLogStreamer().Info("WS", fmt.Sprintf("Client connected (%d total)", len(clients)))

	go cl.writePump()
	go cl.readPump()
}

func (c *client) readPump() {
	defer func() {
		c.conn.Close()
		clientsMu.Lock()
		delete(clients, c)
		clientsMu.Unlock()
		fmt.Printf("[WS] Client disconnected (%d remaining)\n", len(clients))
	}()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var req wsMessage
		if json.Unmarshal(msg, &req) != nil {
			continue
		}

		if req.Type == "request_logs" {
			logStreamer := logging.GetLogStreamer()
			logs := logStreamer.GetRecentLogs(50)
			data, _ := json.Marshal(wsMessage{
				Type: "log_history",
				Data: logs,
			})
			c.send <- data
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, msg)
		case <-ticker.C:
			c.conn.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func BroadcastLog(level, component, message string) {
	data, err := json.Marshal(wsMessage{
		Type: "log_message",
		Data: map[string]string{
			"timestamp": time.Now().Format("15:04:05"),
			"level":     level,
			"component": component,
			"message":   message,
		},
	})
	if err != nil {
		return
	}

	clientsMu.RLock()
	defer clientsMu.RUnlock()

	for cl := range clients {
		select {
		case cl.send <- data:
		default:
		}
	}
}

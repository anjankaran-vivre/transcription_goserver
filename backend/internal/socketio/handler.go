package socketio

import (
	"fmt"
	"time"

	socketio "github.com/googollee/go-socket.io"

	"transcription-goserver/internal/logging"
)

var Server *socketio.Server

func InitSocketIO() *socketio.Server {
	Server = socketio.NewServer(nil)

	Server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		s.Join("all")
		fmt.Printf("[SocketIO] Client connected: %s\n", s.ID())
		logging.GetLogStreamer().Info("SocketIO", fmt.Sprintf("Client %s connected", s.ID()))
		return nil
	})

	Server.OnEvent("/", "request_logs", func(s socketio.Conn, msg string) {
		fmt.Printf("[SocketIO] Client %s requested logs\n", s.ID())
		logStreamer := logging.GetLogStreamer()
		logs := logStreamer.GetRecentLogs(50)

		formattedLogs := make([]map[string]interface{}, 0)
		for _, l := range logs {
			formattedLogs = append(formattedLogs, map[string]interface{}{
				"timestamp": l.Timestamp,
				"level":     l.Level,
				"component": l.Component,
				"message":   l.Message,
			})
		}

		s.Emit("log_history", formattedLogs)
		fmt.Printf("[SocketIO] Sent %d logs to %s\n", len(formattedLogs), s.ID())
	})

	Server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Printf("[SocketIO] Error: %v\n", e)
	})

	Server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Printf("[SocketIO] Client disconnected: %s\n", s.ID())
		logging.GetLogStreamer().Info("SocketIO", fmt.Sprintf("Client %s disconnected", s.ID()))
	})

	go Server.Serve()

	return Server
}

func BroadcastLog(level, component, message string) {
	Server.BroadcastToRoom("/", "all", "log_message", map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"component": component,
		"message":   message,
	})
}

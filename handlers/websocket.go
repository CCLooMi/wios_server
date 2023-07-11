package handlers

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func HandleWebSocket(db *sql.DB) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		// 升级HTTP连接为WebSocket连接
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Failed to upgrade connection to WebSocket:", err)
			return
		}
		defer conn.Close()

		// 处理WebSocket消息
		for {
			// 读取消息
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message from WebSocket:", err)
				break
			}

			// 处理消息
			log.Println("Received message:", string(message))

			// 发送消息
			err = conn.WriteMessage(websocket.TextMessage, []byte("Received your message"))
			if err != nil {
				log.Println("Failed to send message to WebSocket:", err)
				break
			}
		}
	}
}

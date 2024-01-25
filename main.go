package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)

// メッセージブロードキャストチャネル
var broadcast = make(chan Message)

type Message struct {
	Type    int
	Message []byte
}

func main() {
	log.Println("Websocket App start.")

	r := gin.Default()
	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	r.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "templates/index.html")
	})

	r.GET("/ws", func(ctx *gin.Context) {
		conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("Failed to set websocket upgrade: %+v\n", err)
			return
		}
		clients[conn] = true
		for {
			t, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			broadcast <- Message{Type: t, Message: msg}
		}
	})
	go handleMessages()

	r.Run()

	fmt.Println("Websocket App End.")
}

func handleMessages() {
	for {
		message := <-broadcast
		for client := range clients {
			err := client.WriteMessage(message.Type, message.Message)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

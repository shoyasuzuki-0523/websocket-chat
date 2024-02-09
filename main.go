package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

type Message struct {
	Type    int    `json:"type"`
	User    string `json:"user"`
	Message string `json:"message"`
}

type Request struct {
	Header map[string]string `json:"HEADERS"`
	User   string            `json:"user"`
	Input  string            `json:"input"`
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	r.GET("/", func(ctx *gin.Context) {
		rand.New(rand.NewSource(time.Now().UnixNano()))

		ctx.HTML(
			http.StatusOK,
			"index.html",
			gin.H{
				"name": fmt.Sprintf("Guest-%03d", rand.Intn(100)),
			},
		)
	})

	r.GET("/ws", func(ctx *gin.Context) {
		conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("Failed to set websocket upgrade: %+v\n", err)
			return
		}

		clients[conn] = true

		for {
			t, raw, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var req Request
			if err := json.Unmarshal(raw, &req); err != nil {
				fmt.Println(err)
			}

			msg := fmt.Sprintf(
				"<div hx-swap-oob=\"beforeend:#chat\"><span>%s[%s]> %s</span></br></div>",
				time.Now().Format("15:04:05"),
				req.User,
				req.Input,
			)

			broadcast <- Message{Type: t, User: req.User, Message: msg}
		}
	})
	go handleMessages()

	r.Run()
}

func handleMessages() {
	for {
		message := <-broadcast
		for client := range clients {
			if err := client.WriteMessage(message.Type, []byte(message.Message)); err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

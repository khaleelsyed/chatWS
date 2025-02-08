package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	HandshakeTimeout: 5 * time.Second,
}

type ChatServer struct {
	port     int
	sessions map[uuid.UUID]*UserSession
}

func (s *ChatServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS Upgrade error", err)
		return
	}

	uid := uuid.New()
	log.Println("new WS connection:", uid)
	userSession := newUserSesssion(uid, conn, s)
	s.sessions[uid] = &userSession

	go userSession.readLoop()
}

func (s *ChatServer) startHTTP() {
	log.Printf("Starting server on port %d", s.port)
	go func() {
		http.HandleFunc("/ws", s.handleWS)
		err := http.ListenAndServeTLS(fmt.Sprintf(":%d", s.port), "server.crt", "server.key", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func newChatServer(port int) *ChatServer {
	return &ChatServer{
		port:     port,
		sessions: make(map[uuid.UUID]*UserSession, 2),
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	HandshakeTimeout: 5 * time.Second,
}

type UserSession struct {
	uid        uuid.UUID
	conn       *websocket.Conn
	chatserver *ChatServer
}

type ChatServer struct {
	port     int
	sessions map[uuid.UUID]UserSession
}

func (us *UserSession) readLoop() {
	defer func() {
		log.Println("Client Disconnected:", us.uid)
		delete(us.chatserver.sessions, us.uid)
		us.conn.Close()
	}()
}

func (s *ChatServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS Upgrade error", err)
		return
	}

	uid := uuid.New()
	log.Println("Client attempting connection:", uid)
	userSession := newUserSesssion(uid, conn, s)
	s.sessions[uid] = userSession

	go userSession.readLoop()
}

func newUserSesssion(uid uuid.UUID, conn *websocket.Conn, server *ChatServer) UserSession {
	return UserSession{
		uid:        uid,
		conn:       conn,
		chatserver: server,
	}
}

func (s *ChatServer) startHTTP() {
	log.Printf("Starting server on port %d", s.port)
	go func() {
		http.HandleFunc("/ws", s.handleWS)
		http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
	}()
}

func newChatServer(port int) *ChatServer {
	return &ChatServer{
		port:     port,
		sessions: make(map[uuid.UUID]UserSession, 2),
	}
}

func main() {
	port := getPort()

	chatServer := newChatServer(port)
	chatServer.startHTTP()
	select {}
}

func getPort() int {
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	return port
}

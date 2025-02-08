package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/khaleelsyed/chatWS/types"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	HandshakeTimeout: 5 * time.Second,
}

var globalValidator = validator.New()

type UserSession struct {
	uid        uuid.UUID
	Name       string
	conn       *websocket.Conn
	chatserver *ChatServer
}

type ChatServer struct {
	port     int
	sessions map[uuid.UUID]*UserSession
}

func (us *UserSession) handleMessage(msg *types.Message) {
	switch msg.Type {
	case "name_change":
		us.Name = string(msg.Body)
		return
	case "message":
		log.Println("Not ready to handle actual messages yet!")
		return
	default:
		log.Panic(errors.New("handleMessage called with bad type"))
		return
	}
}

func (us *UserSession) readLoop() {
	defer func() {
		log.Println("Client Disconnected:", us.uid)
		delete(us.chatserver.sessions, us.uid)
		us.conn.Close()
	}()
	var msg types.Message
	for {
		var err error
		if err = us.conn.ReadJSON(&msg); err != nil {
			us.notifyError(err)
			continue
		}

		if err = globalValidator.Struct(msg); err != nil {
			us.notifyError(err)
			continue
		}

		go us.handleMessage(&msg)
	}
}

func (us *UserSession) notifyError(err error) {
	log.Println("client", us.uid, "read error: ", err)
	if wsWriteErr := us.conn.WriteMessage(websocket.TextMessage, []byte("bad message")); wsWriteErr != nil {
		log.Println(wsWriteErr)
	}
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

package main

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/khaleelsyed/chatWS/types"
)

var globalValidator = validator.New()

type UserSession struct {
	uid        uuid.UUID
	Name       string
	conn       *websocket.Conn
	chatserver *ChatServer
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

func newUserSesssion(uid uuid.UUID, conn *websocket.Conn, server *ChatServer) UserSession {
	return UserSession{
		uid:        uid,
		conn:       conn,
		chatserver: server,
	}
}

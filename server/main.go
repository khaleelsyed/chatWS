package main

import (
	"log"
	"os"
	"strconv"
)

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

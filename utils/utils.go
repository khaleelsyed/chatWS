package utils

import (
	"log"
	"os"
	"strconv"
)

func getPort() int {
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	return port
}

package main

import (
	"lab1/internal/api"
	"log"
)

func main() {
	log.Println("Application starting...")

	// Запускаем сервер из пакета api
	if err := api.StartServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Application stopped.")
}

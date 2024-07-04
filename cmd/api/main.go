package main

import (
	"log"

	"github.com/hsit18/golang-app/internal/customhttphandler"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("API Services... ")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment variable ", err)
	}

	go customhttphandler.NewServer()
}

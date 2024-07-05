package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	customhttphandler "github.com/hsit18/golang-app/internal/customHttpHandler"
	muxhttphandler "github.com/hsit18/golang-app/internal/muxHttpHandler"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("API Services... ")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment variable ", err)
	}

	go customhttphandler.StartServer()
	go muxhttphandler.NewServer()
	ever := true
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGBUS)
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)

	for ever {
		select {
		case <-sigChan:
			log.Println("Received termination signal, waiting for 10 seconds before shutting down.")
			ever = false
			break
		}
	}
	defer shutdownRelease()
	shutdownServer(shutdownCtx)
	log.Println("Graceful shutdown complete.")
}

func shutdownServer(shutdownCtx context.Context) {
	if err := customhttphandler.StopServer(shutdownCtx); err != nil {
		log.Fatalf("custom HTTP shutdown error: %v", err)
	}
	if err := muxhttphandler.StopServer(shutdownCtx); err != nil {
		log.Fatalf("MUX HTTP shutdown error: %v", err)
	}
}

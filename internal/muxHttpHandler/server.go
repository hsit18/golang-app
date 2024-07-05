package muxhttphandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var myServer *http.Server

func NewServer() {
	myServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("MUX_HTTP_PORT")),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	router := mux.NewRouter()
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
	myServer.Handler = router
	if err := myServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}
	log.Println("Stopped serving new connections.")
}

func StopServer(shutdownCtx context.Context) error {
	log.Println("Shutting down the MUX server...")
	return myServer.Shutdown(shutdownCtx)
}

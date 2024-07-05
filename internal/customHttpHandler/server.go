package customhttphandler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var myServer *http.Server

func StartServer() {

	myServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("HTTP_PORT")),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	myServer.ListenAndServe()
}

func StopServer(shutdownCtx context.Context) error {
	log.Println("Shutting down the server...")
	return myServer.Shutdown(shutdownCtx)
}

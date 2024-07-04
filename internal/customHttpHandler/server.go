package customhttphandler

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func NewServer() *http.Server {
	myServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("HTTP_PORT")),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	myServer.ListenAndServe()
	return myServer
}

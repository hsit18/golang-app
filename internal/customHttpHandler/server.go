package customhttphandler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

var myServer *http.Server

func StartServer() {

	myServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("HTTP_PORT")),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	type Todo struct {
		Id      int
		Message string
	}

	data := map[string][]Todo{
		"Todos": {
			Todo{Id: 1, Message: "Hello Todo"},
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		templ := template.Must(template.ParseFiles("public/index.html"))
		templ.Execute(w, data)
	})
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	myServer.ListenAndServe()
}

func StopServer(shutdownCtx context.Context) error {
	log.Println("Shutting down the server...")
	return myServer.Shutdown(shutdownCtx)
}

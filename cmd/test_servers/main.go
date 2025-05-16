package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <port>")
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid port number")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success response"))
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Slow response"))
	})

	mux.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	mux.HandleFunc("/forbidden", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	})

	mux.HandleFunc("/bad-request", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	})

	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Received POST request"))
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting test server on port %d", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

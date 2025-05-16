package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {
	servers := []struct {
		port     string
		status   int
		response string
	}{
		{port: "9000", status: http.StatusOK, response: "Server 1: OK"},
		{port: "9001", status: http.StatusOK, response: "Server 2: OK"},
		{port: "9002", status: http.StatusInternalServerError, response: "Server 3: Error"},
		{port: "9003", status: http.StatusOK, response: "Server 4: OK"},
		{port: "9004", status: http.StatusOK, response: "Server 5: OK"},
	}

	var wg sync.WaitGroup
	for _, s := range servers {
		wg.Add(1)
		go func(server struct {
			port     string
			status   int
			response string
		}) {
			defer wg.Done()

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				log.Printf("Server on port %s received request: %s %s", server.port, r.Method, r.URL.Path)
				w.WriteHeader(server.status)
				w.Write([]byte(server.response))
			})

			addr := fmt.Sprintf(":%s", server.port)
			log.Printf("Starting server on port %s with status %d", server.port, server.status)
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Printf("Error starting server on port %s: %v", server.port, err)
			}
		}(s)
	}

	log.Println("All test servers are running")
	wg.Wait()
}

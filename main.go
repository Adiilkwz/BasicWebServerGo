package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	mu         sync.Mutex
	data       map[string]string
	requests   int
	shutDownCh chan struct{}
}

func NewServer() *Server {
	return &Server{
		data:       make(map[string]string),
		shutDownCh: make(chan struct{}),
	}
}

type RequestPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Server) postDataHandler(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if payload.Key == "" {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.data[payload.Key] = payload.Value
	s.requests++
	s.mu.Unlock()

	fmt.Printf("Saved to map: Key=%s, Value=%s\n", payload.Key, payload.Value)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Stored Key: %s", payload.Key)
}

func (s *Server) getDataHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()

	dataCopy := make(map[string]string)
	for k, v := range s.data {
		dataCopy[k] = v
	}

	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dataCopy); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) deleteDataHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		http.Error(w, "Key is not found", http.StatusNotFound)
		return
	}

	delete(s.data, key)
	s.requests++

	fmt.Printf("Key %s was deleted\n", key)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deleted Key: %s", key)
}

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := s.requests

	response := map[string]int{
		"total_requests": count,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
	}
}

func (s *Server) startBackgroundWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			count := s.requests
			size := len(s.data)
			s.mu.Unlock()

			fmt.Printf("[Worker Log] Requests: %d | Database Size: %d\n", count, size)
		case <-s.shutDownCh:
			fmt.Println("Worker Stopped")
			return
		}
	}
}

func main() {
	server := NewServer()
	mux := http.NewServeMux()

	mux.HandleFunc("POST /data", server.postDataHandler)
	mux.HandleFunc("GET /data", server.getDataHandler)
	mux.HandleFunc("DELETE /data/{key}", server.deleteDataHandler)
	mux.HandleFunc("GET /stats", server.statsHandler)

	go server.startBackgroundWorker()

	fmt.Println("Server starting on :8000")

	if err := http.ListenAndServe(":8000", mux); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

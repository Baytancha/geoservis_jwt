package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
)

func TestMainfunc(t *testing.T) {
	go func() {
		main()
	}()
	time.Sleep(2 * time.Second)
	t.Log("main finished")
}

func TestReverseProxy_proxy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api", nil)
	w := httptest.NewRecorder()
	r := chi.NewRouter()
	proxy := NewReverseProxy("hugo_task", "1313")
	r.Use(proxy.ReverseProxy)
	r.Get("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from API"))
	}))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	if w.Body.String() != "Hello from API" {
		t.Errorf("Expected 'Hello from API', got %s", w.Body.String())
	}
}

func TestReverseProxy_target(t *testing.T) {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":1313",
		Handler: mux,
	}
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("TASK LIST"))
	})
	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/tasks", nil)
	w := httptest.NewRecorder()
	r := chi.NewRouter()
	proxy := NewReverseProxy("localhost", "1313")
	r.Use(proxy.ReverseProxy)
	r.Get("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from API"))
	}))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	if w.Body.String() != "TASK LIST" {
		t.Errorf("Expected 'Hello from API', got %s", w.Body.String())
	}
}

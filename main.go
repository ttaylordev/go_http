package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type loggingHandler struct {
	handler http.Handler
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
    rw.status = statusCode
    rw.ResponseWriter.WriteHeader(statusCode)
}

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    
    // Create a custom ResponseWriter to capture the response status
    rw := &responseWriter{ResponseWriter: w}
   
    log.Printf("Request: %s %s", r.Method, r.URL.Path)
    h.handler.ServeHTTP(rw, r)
    
    // Check if the response status is 404 (Not Found)
    if rw.status == http.StatusNotFound {
        log.Printf("File not found: %s", r.URL.Path)
    }
}

func main() {

    // Serve the static files
	dir := http.Dir("./dist/")
	fs := http.FileServer(dir)

	// Wrap the file server handler with logging
	fsWithLogging := &loggingHandler{handler: fs}

	// Create a custom error handler for failed requests
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	})

	// Create a multiplexer for handling different routes
	mux := http.NewServeMux()
	mux.Handle("/", fsWithLogging)         // Serve static files with logging
	mux.Handle("/favicon.ico", fs)        // Serve favicon.ico directly from the file server
	mux.Handle("/api", errorHandler)      // Custom error handling for /api route
	mux.Handle("/api/data", errorHandler) // Custom error handling for /api/data route

	// Create a new HTTP server with the multiplexer
	srv := &http.Server{
		Addr:    ":8018",
		Handler: mux,
	}

	// Start the server in a separate goroutine
	go func() {
		log.Println("Server started on http://localhost:8018")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-quit
	log.Println("Server is shutting down...")

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
    
	// Attempt to gracefully shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully stopped")
}

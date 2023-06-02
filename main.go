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

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Printf("Request: %s %s", r.Method, r.URL.Path)
    h.handler.ServeHTTP(w, r)
}

func main() {



    // Serve the React.js static files
    dir := http.Dir("./dist/")
    // Wrap the file server handler with logging
    fs := &loggingHandler{handler: http.FileServer(dir)}
	
    // Check for errors when creating the file server
    file,err := dir.Open("")

    if err != nil{
        log.Fatal(err)
    }

    file.Close()


    srv := &http.Server {
        Addr: ":8018",
        Handler: fs,

    }

    // Start server in a seperate go routine
    go func() {
		log.Println("Server started on http://localhost:8018")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    // Block until a signal is received
    <-quit
    log.Println("Server is shutting down...")

    // Create a context with a timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Attempt to gracefully shut down the server
    if err:= srv.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }

    log.Println("Server gracefully stopped")


}

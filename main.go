package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

func init() {
	// Set log prefix based on application name from args
	appName := filepath.Base(os.Args[0]) // Removes leading "./" if present
	log.SetPrefix("[" + appName + "] ")
	//log.SetFlags(0) // Optional: removes default date and time from log output
}

func main() {
	// Set up module paths
	moduleCache, goPath, err := setupModulePaths()
	if err != nil {
		log.Fatalf("Failed to set up module paths: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(services))

	// Create a context that is canceled on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start services
	for i := range services {
		if err := startService(ctx, &wg, &services[i], moduleCache, goPath); err != nil {
			log.Fatalf("Failed to start service %s: %v", services[i].Path, err)
		}
	}

	// Set up HTTP handlers
	setupHTTPHandlers()

	// Extract server port from environment variable, default to 8080 if missing
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	// Create an HTTP server
	srv := &http.Server{
		Addr: ":" + serverPort,
	}

	// Start HTTP server in a goroutine
	go func() {

		// // Handle requests for the root path and serve the specific HTML file
		// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 	http.ServeFile(w, r, staticDir+"/WebRunner.html")
		// })

		log.Printf("Starting server on port: %s", serverPort)
		log.Printf("Usage: localhost:8081/gen \n")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start driver HTTP server: %v", err)
		}
	}()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigChan
	log.Printf("Received signal: %v. Initiating graceful shutdown...", sig)

	// Create a context with timeout for the shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown the HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown error: %v", err)
	} else {
		log.Printf("HTTP server shutdown gracefully\n")
	}

	// Cancel the context to stop child processes
	cancel()

	// Stop all services
	for _, service := range services {
		if service.Cmd != nil && service.Cmd.Process != nil {
			pgid, err := syscall.Getpgid(service.Cmd.Process.Pid)
			if err == nil {
				// Send SIGTERM to the process group
				if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
					log.Printf("Failed to send SIGTERM to service %s: %v", service.Path, err)
				}
			} else {
				log.Printf("Failed to get pgid for service %s: %v", service.Path, err)
			}
		}
	}

	// Wait for all services to stop
	wg.Wait()
	log.Printf("All services stopped. Exiting.\n")
}

// setupModulePaths sets up paths for GOMODCACHE and GOPATH and ensures the cache directories exist
func setupModulePaths() (string, string, error) {
	// Set up paths for GOMODCACHE and GOPATH if they're not defined
	moduleCache := "/tmp/gomodcache"
	goPath := "/tmp/gopath"

	// Ensure the cache directories exist
	if err := os.MkdirAll(moduleCache, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("failed to create module cache directory: %w", err)
	}
	if err := os.MkdirAll(goPath, os.ModePerm); err != nil {
		return "", "", fmt.Errorf("failed to create GOPATH directory: %w", err)
	}

	return moduleCache, goPath, nil
}

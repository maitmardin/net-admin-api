package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"net-admin/internal/server"
)

const DEFAULT_PORT            = 8080
const DEFAULT_VLAN_STORE_PATH = "vlans.json"

func main() {
	port := DEFAULT_PORT
	if portStr := os.Getenv("PORT"); portStr != "" {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("invalid API server port: %v", err)
		}
	}

	vlanStorePath := os.Getenv("VLAN_STORE_PATH")
	if vlanStorePath == "" {
		vlanStorePath = DEFAULT_VLAN_STORE_PATH
	}

	server, err := server.NewServer(port, vlanStorePath)
	if err != nil {
		log.Fatalf("failed to create API server: %v", err)
	}

	shutdownDone := make(chan bool, 1)
	go gracefulShutdown(server, shutdownDone)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("API server error: %s", err)
	}
	log.Printf("Started API server at %v", server.Addr)

	<-shutdownDone
	log.Println("API server shutdown complete")
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("Shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown by not diverting the signals any more

	// Allow 5 seconds to finish ongoing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	done <- true
}

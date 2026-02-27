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

const (
	defaultPort = "18080"
	configFile  = "openclaw.json"
)

func main() {
	log.Println("OpenClaw Installer Starting...")

	// Detect platform
	platform := DetectPlatform()
	log.Printf("Detected platform: %s/%s", platform.OS, platform.Arch)

	// Create installer instance
	installer := NewInstaller(platform)

	// Create and start web server
	server := NewServer(installer)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		addr := ":" + defaultPort
		log.Printf("Starting web server on http://localhost%s", addr)

		// Auto-open browser
		go func() {
			time.Sleep(500 * time.Millisecond)
			if err := OpenBrowser("http://localhost" + addr); err != nil {
				log.Printf("Failed to open browser: %v", err)
			}
		}()

		if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/go-cloud/health"
)

var (
	listenAddr string
	redisAddr  string
	redisPass  string
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	flag.StringVar(&listenAddr, "listen-addr", getEnv("LISTEN_ADDR", ":5000"), "server listen address")
	flag.StringVar(&redisAddr, "redis-addr", getEnv("REDIS_ADDR", "localhost:6379"), "redis address")
	flag.StringVar(&redisPass, "redis-password", getEnv("REDIS_PASS", ""), "redis password")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	redisHealthChecker := NewRedisHealthChecker(redisAddr, redisPass, log.New(os.Stdout, "redis: ", log.LstdFlags))

	healthz := health.Handler{}
	healthz.Add(redisHealthChecker)

	server := &http.Server{
		Addr:     listenAddr,
		Handler:  &healthz,
		ErrorLog: logger,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", listenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

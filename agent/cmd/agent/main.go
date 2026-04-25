package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubenexus/agent/internal/agent"
)

func main() {
	serverURL := os.Getenv("SERVER_URL")
	clusterToken := os.Getenv("CLUSTER_TOKEN")
	clusterID := os.Getenv("CLUSTER_ID")

	if serverURL == "" || clusterToken == "" {
		log.Fatal("SERVER_URL and CLUSTER_TOKEN environment variables are required")
	}

	cfg := &agent.Config{
		ServerURL:    serverURL,
		ClusterToken: clusterToken,
		ClusterID:    clusterID,
		HeartbeatInterval: 30 * time.Second,
		SyncInterval:     60 * time.Second,
		KubeconfigPath:   os.Getenv("KUBECONFIG"),
	}

	a := agent.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Received shutdown signal, gracefully stopping...")
		cancel()
	}()

	if err := a.Run(ctx); err != nil {
		log.Fatalf("Agent exited with error: %v", err)
	}
	log.Println("Agent stopped")
}

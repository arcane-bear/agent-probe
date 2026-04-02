package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arcane-bear/agent-probe/pkg/checker"
	"github.com/arcane-bear/agent-probe/pkg/config"
)

func main() {
	configPath := flag.String("config", "/etc/agent-probe/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	c := checker.New(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler(c))
	mux.HandleFunc("/readyz", readyzHandler(c))

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("agent-probe listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func healthzHandler(c *checker.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := c.RunAll()
		writeResult(w, result)
	}
}

func readyzHandler(c *checker.Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := c.RunReadiness()
		writeResult(w, result)
	}
}

func writeResult(w http.ResponseWriter, result *checker.Result) {
	w.Header().Set("Content-Type", "application/json")

	status := http.StatusOK
	if result.Status == "unhealthy" {
		status = http.StatusServiceUnavailable
	}

	result.Timestamp = time.Now().UTC().Format(time.RFC3339)

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(result)
}

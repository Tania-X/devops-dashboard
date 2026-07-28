package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9100"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		snapshot, err := monitor.Collect()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	})

	log.Printf("[Agent] 启动，监听端口 :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("[Agent] 启动失败: %v", err)
	}
}

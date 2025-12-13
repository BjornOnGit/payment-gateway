package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Alert represents the alert structure received from services
type Alert struct {
	Service   string                 `json:"service"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

var (
	alerts []Alert
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %v", err)
		http.Error(w, "error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var alert Alert
	if err := json.Unmarshal(body, &alert); err != nil {
		log.Printf("error unmarshaling alert: %v", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	alerts = append(alerts, alert)

	log.Printf("[%s] %s: %s | Details: %v",
		alert.Severity,
		alert.Service,
		alert.Message,
		alert.Details,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"ok","message":"alert received"}`)
}

func alertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(alerts),
		"alerts": alerts,
	})
}

func clearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	alerts = []Alert{}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"ok","message":"alerts cleared"}`)
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/alerts", alertsHandler)
	http.HandleFunc("/clear", clearHandler)

	log.Println("Alert webhook receiver listening on :9000")
	log.Println("  POST /webhook - Receive alerts from services")
	log.Println("  GET  /alerts  - View all received alerts")
	log.Println("  POST /clear   - Clear alert history")

	log.Fatal(http.ListenAndServe(":9000", nil))
}

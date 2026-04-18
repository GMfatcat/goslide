package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"yield":  96.2,
			"line":   r.URL.Query().Get("line"),
			"status": "running",
		})
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"lines": []map[string]any{
				{"name": "Line A", "yield": 96.2},
				{"name": "Line B", "yield": 93.8},
				{"name": "Line C", "yield": 97.1},
			},
		})
	})

	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		now := time.Now().Format("2006-01-02 15:04:05")
		events := []string{"Inspection OK", "Defect detected", "Camera recalibrated", "Model inference", "Yield updated"}
		yields := []string{"95.8%", "96.2%", "97.1%", "93.4%", "96.7%"}
		json.NewEncoder(w).Encode(map[string]any{
			"entries": []string{
				fmt.Sprintf("[%s] %s", now, events[rand.Intn(len(events))]),
				fmt.Sprintf("[%s] Yield: %s", now, yields[rand.Intn(len(yields))]),
				fmt.Sprintf("[%s] Active lines: %d/4", now, 3+rand.Intn(2)),
			},
		})
	})

	fmt.Println("Mock API running on http://localhost:9999")
	http.ListenAndServe(":9999", nil)
}

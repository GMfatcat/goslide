package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		json.NewEncoder(w).Encode(map[string]any{
			"entries": []string{
				"[2026-04-18 10:00:01] System startup",
				"[2026-04-18 10:00:02] Camera initialized",
				"[2026-04-18 10:00:03] Model loaded: SegFormer v2",
				"[2026-04-18 10:00:05] First inspection completed",
				"[2026-04-18 10:00:08] Yield: 96.2%",
			},
		})
	})

	fmt.Println("Mock API running on http://localhost:9999")
	http.ListenAndServe(":9999", nil)
}

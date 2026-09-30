// Command demoserver supplies invented API responses for an account-free VHS demo.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := "127.0.0.1:8643"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           http.HandlerFunc(serveDemo),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	log.Println("invented demo API listening")
	log.Fatal(server.ListenAndServe())
}

func serveDemo(w http.ResponseWriter, r *http.Request) {
	var result any
	switch r.Method + " " + r.URL.Path {
	case "GET /api":
		result = []any{
			map[string]any{"legal": "Invented demo listings; no live job data.", "last_updated": 1790762400},
			map[string]any{"id": "demo-go-01", "position": "Go Platform Engineer", "company": "Demo Orbit Workshop", "tags": []string{"golang", "remote"}, "location": "Worldwide", "salary_min": 110000, "salary_max": 145000, "date": "2026-09-30T10:00:00Z", "epoch": 1790762400, "description": "Build Go services and developer tools for an invented workshop.", "url": "https://jobs.example/demo-go-01"},
			map[string]any{"id": "demo-go-02", "position": "Go Backend Engineer", "company": "Demo Comet Lab", "tags": []string{"golang", "remote"}, "location": "Worldwide", "salary_min": 95000, "salary_max": 125000, "date": "2026-09-29T10:00:00Z", "epoch": 1790676000, "description": "Design APIs and test reliable Go services.", "url": "https://jobs.example/demo-go-02"},
			map[string]any{"id": "demo-design-03", "position": "Product Designer", "company": "Demo Pixel Workshop", "tags": []string{"design", "remote"}, "location": "Worldwide", "salary_min": 80000, "salary_max": 100000, "date": "2026-09-28T10:00:00Z", "epoch": 1790589600, "description": "Shape an invented product's interface.", "url": "https://jobs.example/demo-design-03"},
		}
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Print(err)
	}
}

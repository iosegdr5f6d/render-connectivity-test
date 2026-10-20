
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width">
<title>Render Connectivity Test</title>
<style>
body {
  background: #151720;
  color: #e8eaf0;
  font: 16px Arial, sans-serif;
  max-width: 650px;
  margin: 60px auto;
  padding: 20px;
}
a { color: #9eafff; }
</style>
</head>
<body>
<h1>Render backend is running</h1>
<p>The service has responded to an HTTP request.</p>
<p><a href="/api/health">Open health endpoint</a></p>
</body>
</html>`))
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET required", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"message": "Backend is reachable",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	addr := "0.0.0.0:" + port
	log.Printf("Starting connectivity test on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	targetHost := os.Getenv("TARGET_HOST")
	if targetHost == "" {
		log.Fatal("TARGET_HOST environment variable is required")
	}

	listenPort := os.Getenv("LISTEN_PORT")
	if listenPort == "" {
		listenPort = "80"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "It works!")
			return
		}

		if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			forwardACMEChallenge(w, r, targetHost)
			return
		}

		http.NotFound(w, r)
	})

	addr := ":" + listenPort
	log.Printf("Listening on %s, forwarding ACME challenges to %s", addr, targetHost)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func forwardACMEChallenge(w http.ResponseWriter, r *http.Request, targetHost string) {
	if !strings.Contains(targetHost, ":") {
		targetHost = targetHost + ":80"
	}

	targetURL := fmt.Sprintf("http://%s%s", targetHost, r.URL.Path)

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Host = r.Host

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to %s: %v", targetURL, err)
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

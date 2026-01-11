package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	targetHost := os.Getenv("TARGET_HOST")
	if targetHost == "" {
		slog.Error("TARGET_HOST environment variable is required")
		os.Exit(1)
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
	slog.Info("starting server", "addr", addr, "target_host", targetHost)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func forwardACMEChallenge(w http.ResponseWriter, r *http.Request, targetHost string) {
	if !strings.Contains(targetHost, ":") {
		targetHost = targetHost + ":80"
	}

	targetURL := fmt.Sprintf("http://%s%s", targetHost, r.URL.Path)

	slog.Info("forwarding ACME challenge", "host", r.Host, "path", r.URL.Path, "target", targetURL)

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Host = r.Host

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to forward request", "target", targetURL, "error", err)
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

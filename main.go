package main

import (
	"fmt"
	"net/http"
	"sync"
	"SystemRemoteDevice/project"
)

var (
    mu      sync.Mutex
    running bool
    stopCh  chan struct{}
)

func startHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if running {
		fmt.Fprintln(w, "Monitoring sudah aktif")
		return
	}

	stopCh = make(chan struct{})
	running = true

	go project.MonitoringNetwork(stopCh)
	fmt.Fprintln(w, "Monitoring diaktifkan")
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()

    if !running {
        fmt.Fprintln(w, "Monitoring sudah tidak aktif")
        return
    }

    close(stopCh)
    running = false
    fmt.Fprintln(w, "Monitoring dimatikan")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if running {
		fmt.Fprintln(w, "Status: Monitoring AKTIF")
	} else {
		fmt.Fprintln(w, "Status: Monitoring NONAKTIF")
	}
}

func main() {
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/status", statusHandler)

	fmt.Println("Server jalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
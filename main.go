package main

import (
	"SystemRemoteDevice/project"
	"fmt"
	"net/http"
	"sync"
)

var (
	muMonitor      sync.Mutex
	monitorRunning bool
	monitorStopCh  chan struct{}
)

// === Monitoring Network Handlers ===

// Starts the network monitoring process
func startMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	muMonitor.Lock()
	defer muMonitor.Unlock()

	if monitorRunning {
		fmt.Fprintln(w, "Monitoring Network is already running.")
		return
	}

	monitorStopCh = make(chan struct{})
	monitorRunning = true

	go project.MonitoringNetwork(monitorStopCh)
	fmt.Fprintln(w, "Monitoring Network has been started.")
}

// Stops the network monitoring process
func stopMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	muMonitor.Lock()
	defer muMonitor.Unlock()

	if !monitorRunning {
		fmt.Fprintln(w, "Monitoring Network is not currently running.")
		return
	}

	close(monitorStopCh)
	monitorRunning = false
	fmt.Fprintln(w, "Monitoring Network has been stopped.")
}

// Returns the current status of the monitoring process
func statusMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	muMonitor.Lock()
	defer muMonitor.Unlock()

	if monitorRunning {
		fmt.Fprintln(w, "Monitoring Network Status: ACTIVE")
	} else {
		fmt.Fprintln(w, "Monitoring Network Status: INACTIVE")
	}
}

func startSysInfoHandler(w http.ResponseWriter, r *http.Request) {
	go project.GetSystemInformation()
	fmt.Fprintln(w, "System Information retrieval has been started.")
}
func startCheckSystemHasError(w http.ResponseWriter, r *http.Request) {
	go project.CheckSystemHasError()
	fmt.Fprintln(w, "CheckSystemHas Error retrieval has been started.")
}

func main() {
	// Route for Monitoring Network
	http.HandleFunc("/project/monitoring-network/start", startMonitoringHandler)
	http.HandleFunc("/project/monitoring-network/status", statusMonitoringHandler)
	http.HandleFunc("/project/monitoring-network/stop", stopMonitoringHandler)

	// Route for System Information
	http.HandleFunc("/project/get-system-information/start", startSysInfoHandler)
	http.HandleFunc("/project/check-system-has-error/start", startCheckSystemHasError)

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

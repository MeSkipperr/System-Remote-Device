package handlers

import (
	"SystemRemoteDevice/project"
	"sync"
	"net/http"
	"fmt"
)
var (
	muMonitor      sync.Mutex
	monitorRunning bool
	monitorStopCh  chan struct{}
)

// === Monitoring Network Handlers ===

// Starts the network monitoring process
func StartMonitoringHandler(w http.ResponseWriter, r *http.Request) {
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
func StopMonitoringHandler(w http.ResponseWriter, r *http.Request) {
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
func StatusMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	muMonitor.Lock()
	defer muMonitor.Unlock()

	if monitorRunning {
		fmt.Fprintln(w, "Monitoring Network Status: ACTIVE")
	} else {
		fmt.Fprintln(w, "Monitoring Network Status: INACTIVE")
	}
}
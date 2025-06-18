package main

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/project"
	"SystemRemoteDevice/utils/schedule"
	"fmt"
	"net/http"
	"sync"

	"github.com/robfig/cron/v3"
)

var (
	muMonitor      sync.Mutex
	monitorRunning bool
	monitorStopCh  chan struct{}
)

type scheduleConfig struct {
	RemoveYoutubeData       	string      `json:"removeYoutubeData"`
	GetUptimeADBDevices       	string      `json:"getUptimeADBDevices"`
	RestartComputer       		string      `json:"restartComputer"`
	ClearLogMonitoring       	string      `json:"clearLogMonitoring"`
}
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
func startGetUptimeADB(w http.ResponseWriter, r *http.Request) {
	go project.GetUptimeADB()
	fmt.Fprintln(w, "Get uptime device ADB process initiated.")
}

func main() {
	
	go func() {
		scheduleConfig, errLoadJson := config.LoadJSON[scheduleConfig]("utils/schedule/schedule.json")
		if errLoadJson != nil {	
			fmt.Println("Failed to load config from json", errLoadJson)
			return 
		} 
		c := cron.New()

		c.AddFunc(scheduleConfig.ClearLogMonitoring	, schedule.ClearMonitoringLog)
		c.AddFunc(scheduleConfig.RestartComputer	, schedule.RestartComputer)
		c.AddFunc(scheduleConfig.GetUptimeADBDevices, project.GetUptimeADB)
		c.AddFunc(scheduleConfig.RemoveYoutubeData	, project.RemoveYouTubeData)

		c.Start()
	}()

	// Route for Monitoring Network
	http.HandleFunc("/project/monitoring-network/start", startMonitoringHandler)
	http.HandleFunc("/project/monitoring-network/status", statusMonitoringHandler)
	http.HandleFunc("/project/monitoring-network/stop", stopMonitoringHandler)

	// Route for System Information
	http.HandleFunc("/project/get-system-information/start", startSysInfoHandler)
	http.HandleFunc("/project/check-system-has-error/start", startCheckSystemHasError)
	http.HandleFunc("/project/get-uptime-ADB-devices/start", startGetUptimeADB)

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

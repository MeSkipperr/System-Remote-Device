package main

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/handlers"
	"SystemRemoteDevice/project"
	"SystemRemoteDevice/routes"
	"SystemRemoteDevice/utils/schedule"
	"fmt"
	"net/http"

	"github.com/robfig/cron/v3"
)



type scheduleConfig struct {
	RemoveYoutubeData       	string      `json:"removeYoutubeData"`
	GetUptimeADBDevices       	string      `json:"getUptimeADBDevices"`
	RestartComputer       		string      `json:"restartComputer"`
	ClearLogMonitoring       	string      `json:"clearLogMonitoring"`
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
		c.AddFunc(scheduleConfig.ClearLogMonitoring	, schedule.DeleteLogFiles)
		c.AddFunc(scheduleConfig.RestartComputer	, schedule.RestartComputer)
		c.AddFunc(scheduleConfig.GetUptimeADBDevices, project.GetUptimeADB)
		c.AddFunc(scheduleConfig.RemoveYoutubeData	, project.RemoveYouTubeData)

		c.Start()
	}()
	//auto running function at first time

	go func() {
		project.RemoveYouTubeData();
		project.GetSystemInformation()
		project.CheckSystemHasError()	
		handlers.AutoStartMonitoringNetwork()
	}()
		
	r := routes.RegisterRoutes()

	fmt.Println("Server running on http://localhost:8000")
	http.ListenAndServe(":8000", r) 
}

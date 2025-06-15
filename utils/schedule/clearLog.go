package schedule

import (
	"SystemRemoteDevice/utils"
	"fmt"
	"os"
	"path/filepath"
	"SystemRemoteDevice/config"
)	

type monitoringNetworkType struct {
	Times 		int				`json:"times"`
	Runtime  	int				`json:"runtime"`
	DeviceType 	[]string		`json:"deviceType"`
	LogPath		string			`json:"logPath"`
}

func ClearMonitoringLog() {
	fmt.Println("Running clearMonitoringLog:", utils.GetCurrentTimeFormatted())
	conf, err := config.LoadJSON[monitoringNetworkType]("config/monitoring-network.json")

	if err != nil {	
		fmt.Println("Failed to load config from json", err)
		return 
	}
	logFolder := conf.LogPath;

	filepath.Walk(logFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fmt.Println("Deleting:", path)
			os.Remove(path)
		}
		return nil
	})

}
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
	OutputPath		string			`json:"outputPath"`
}

func ClearMonitoringLog() {
	fmt.Println("Running clearMonitoringLog:", utils.GetCurrentTimeFormatted())
	conf, err := config.LoadJSON[monitoringNetworkType]("config/monitoring-network.json")

	if err != nil {	
		fmt.Println("Failed to load config from json", err)
		return 
	}
	logFolder := conf.OutputPath;

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

// DeleteLogFiles deletes all .log files in the ./log directory.
func DeleteLogFiles() {
	logDir := "./log"

	files, _ := filepath.Glob(filepath.Join(logDir, "*.log"))

	if len(files) == 0 {
		fmt.Println("No .log files found in:", logDir)
		return
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			fmt.Printf("Failed to delete file %s: %v\n", file, err)
		} else {
			fmt.Printf("Deleted: %s\n", file)
		}
	}
}
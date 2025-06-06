package project

import (
	"fmt"
	"strings"
	"time"
	"SystemRemoteDevice/utils"
	"SystemRemoteDevice/config"
)


type Device struct {
	Name        string `json:"name"`
	IPAddress   string `json:"ipAddress"`
	Device      string `json:"device"`
	Error       bool   `json:"error"`
	Description string `json:"description"`
}

func MonitoringNetwork(stopChan chan struct{}) {
	fmt.Println("Monitoring network started")

	times := config.MonitoringNetwork.Times

	ticker := time.NewTicker(time.Duration(config.MonitoringNetwork.Runtime) * time.Second)
	defer ticker.Stop()

	devices := []Device{
		{
			Name:        "Network",
			IPAddress:   "8.8.8.8",
			Device:      "Google",
			Error:       false,
			Description: "Public Ip",
		},
		{
			Name:        "Server 1",
			IPAddress:   "10.10.10.2",
			Device:      "Server HP",
			Error:       false,
			Description: "Web server",
		},
	}

	for {
		select {
		case <-ticker.C:
			for i := range devices {
				dev := &devices[i]
				replies, err := utils.PingDivice(dev.IPAddress, times)
				if err != nil {
					fmt.Printf("Time: %s | Name: %s | IP: %s | Device: %s | Lost Percent: 100%%\n",
						utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device)
					continue
				}

				lines := strings.Split(replies, "\n")
				// Amanin supaya gak index out of range
				var linesResArray []string
				if len(lines) >= 2+times {
					linesResArray = lines[2 : 2+times]
				} else {
					linesResArray = lines
				}

				lostCount := 0
				for _, r := range linesResArray {
					if strings.Contains(r, "Destination Host Unreachable") ||
						strings.Contains(r, "Request timed out") ||
						strings.Contains(r, "100% packet loss") ||
						strings.Contains(r, "Name or service not known") ||
						strings.Contains(r, "could not find host") {
						lostCount++
					}
				}
				lostPercent := (float64(lostCount) / float64(times)) * 100

				if lostPercent == 100 && !dev.Error {
					// Kirim email error bisa di sini
					dev.Error = true
				} else if lostPercent < 100 && dev.Error {
					// Kirim email recovery bisa di sini
					dev.Error = false
				}
				fmt.Printf("Time: %s | Name: %s | IP: %s | Device: %s | Lost Percent: %.2f%%\n",
					utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device, lostPercent)
			}
		case <-stopChan:
			fmt.Println("Monitoring network stopped")
			return
		}
	}
}


// func MonitoringNetwork() {
// 	fmt.Println("=========")
// 	fmt.Println("Network ping device")
// 	fmt.Println("=========")

// 	times := config.MonitoringNetwork.Times

	
	// devices := []Device{
	// 	{
	// 		Name:        "Network",
	// 		IPAddress:   "8.8.8.8",
	// 		Device:      "Google",
	// 		Error:       false,
	// 		Description: "Public Ip",
	// 	},
	// 	{
	// 		Name:        "Server 1",
	// 		IPAddress:   "10.10.10.2",
	// 		Device:      "Server HP",
	// 		Error:       false,
	// 		Description: "Web server",
	// 	},
	// }
// 	ticker := time.NewTicker(time.Duration(config.MonitoringNetwork.Runtime) * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		<-ticker.C
// 		for _, dev := range devices {
// 			isError := dev.Error

// 			replies, err := utils.PingDivice(dev.IPAddress, times)
// 			if err != nil {
// 			fmt.Printf("Time: %s | Name: %s | IP: %s | Device: %s | Lost Percent: 100%%\n",
// 				utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device)
// 			}
// 			lines := strings.Split(replies, "\n")
// 			linesResArray := lines[2 : 2+times]

// 			lostCount := 0
// 			for _, r := range linesResArray {
// 				if strings.Contains(r, "Destination Host Unreachable") ||
// 					strings.Contains(r, "Request timed out") ||
// 					strings.Contains(r, "100% packet loss") ||
// 					strings.Contains(r, "Name or service not known") ||
// 					strings.Contains(r, "could not find host") {
// 					lostCount++
// 				}
// 			}
// 			lostPercent := (float64(lostCount) / float64(times)) * 100

// 			if lostPercent == 100 && !isError {
// 				//send email
// 				isError = true
// 			} else if lostPercent == 50 && isError {
// 				//send email
// 				isError = false
// 			}
// 			fmt.Printf("Time: %s | Name: %s | IP: %s | Device: %s | Lost Percent: %.2f%%\n",
// 				utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device, lostPercent)
// 		}
// 	}
// }

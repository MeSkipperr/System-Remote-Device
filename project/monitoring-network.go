package project

import (
	"fmt"
	"sync"
	"time"
		"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"SystemRemoteDevice/template/email"
	"database/sql"
	_ "modernc.org/sqlite"
	"strings"
)

type monitoringNetworkType struct {
	Times 		int				`json:"times"`
	Runtime  	int				`json:"runtime"`
	DeviceType 	[]string		`json:"deviceType"`
	LogPath		string			`json:"logPath"`
	OutputPath		string		`json:"outputPath"`
}


func updateError(dev models.DeviceType, errorStatus bool, ) (bool, string) {
	currentTime := time.Now()

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		return false, fmt.Sprintf("Database open error: %v", err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`UPDATE devices SET error = ?, down_time = ? WHERE id = ?`)
	if err != nil {
		return false, fmt.Sprintf("Prepare statement error: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(errorStatus, currentTime, dev.ID)
	if err != nil {
		return false, fmt.Sprintf("Exec update error: %v", err)
	}

	return true, "Success: Error status updated for device"
}

func statusChecking(dev models.DeviceType, lostPercent float64,outputPath string) {
	conf, err := config.LoadJSON[monitoringNetworkType]("config/monitoring-network.json")

	if err != nil {	
		fmt.Println("Failed to load config from json", err)
		return 
	}

	logText := fmt.Sprintf("Time: %s | Name: %s | IP: %s | Device: %s | Lost Percent: %.2f%%\n",
		utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device, lostPercent)

	errWriteTxt := utils.WriteToTXT(outputPath+dev.Name+".txt", logText, true)

	if errWriteTxt != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Write Log", fmt.Sprintf("Failed to write file .txt: %v", errWriteTxt)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)	
		}
		return
	} 

	if lostPercent == 100 && !dev.Error {
		setError, errMsg := updateError(dev, true)

		if !setError {
			if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Update Error Status", fmt.Sprintf("Failed to update error status on device: %v", errMsg)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
		} 

		email := models.EmailStructure{
			EmailData: models.EmailData{
				Subject:       "Device Down Notification",
				BodyTemplate:  email.ErrorDeviceEmail(dev),
				FileAttachment: []string{},
			},
		}

		success, message := utils.SendEmail(email)

		if success {
			if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Email", fmt.Sprintf("Email sent successfully: %s", message)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
		} else {
			if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Email", fmt.Sprintf("Failed to send email: %s", message)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
		}

	} else if lostPercent == 0 && dev.Error {
		// Kirim email recovery bisa di sini
		setError, errMsg := updateError(dev, false)

		if !setError {
			if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Update Error Status", fmt.Sprintf("Failed to update error status on device: %v", errMsg)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
		} 

		email := models.EmailStructure{
			EmailData: models.EmailData{
				Subject:       "Device Recovery Notification",
				BodyTemplate:  email.RecoveryDeviceEmail(dev),
				FileAttachment: []string{},
			},
		}

		success, message := utils.SendEmail(email)

		if success {
			if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Email", fmt.Sprintf("Email sent successfully: %s", message)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
		} else {
			if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Email", fmt.Sprintf("Failed to send email: %s", message)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
		}

	}
}



func startDeviceLoop(dev models.DeviceType, conf  monitoringNetworkType, stopChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(conf.Runtime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		default:
				times := conf.Times

				replies, err := utils.PingDevice(dev.IPAddress, times)
				if err != nil {
					statusChecking(dev, 100,conf.OutputPath)
					if errLog := utils.WriteFormattedLog(
						conf.LogPath,
						"INFO",
						"Ping Device",
						fmt.Sprintf("Name: %s | IP: %s | Device: %s | Lost Percent: %.2f%%", dev.Name, dev.IPAddress, dev.Device, 100.0),
					); errLog != nil {
						fmt.Printf("Failed to write log: %v\n", errLog)
					}

					continue
				}

				lines := strings.Split(replies, "\n")
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

				statusChecking(dev, lostPercent,conf.OutputPath)
				if errLog := utils.WriteFormattedLog(
					conf.LogPath,
					"INFO",
					"Ping Device",
					fmt.Sprintf("Name: %s | IP: %s | Device: %s | Lost Percent: %.2f%%", dev.Name, dev.IPAddress, dev.Device, lostPercent),
				); errLog != nil {
					fmt.Printf("Failed to write log: %v\n", errLog)
				}
			select {
			case <-ticker.C:
				// Lanjut ke iterasi berikutnya
			case <-stopChan:
				return
			}
		}
	}
}


func MonitoringNetwork(stopChan chan struct{}) {
	var wg sync.WaitGroup

	conf, err := config.LoadJSON[monitoringNetworkType]("config/monitoring-network.json")

	if err != nil {	
		fmt.Println("Failed to load config from json", err)
		return 
	}

	if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Monitoring Network",fmt.Sprintf("Monitoring network started at %s",utils.GetCurrentTimeFormatted())); errLog != nil {
		fmt.Printf("Failed to write log: %v\n", errLog)
	}


	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf("Error connecting to database: %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		panic(err)
	}
	defer db.Close()

	deviceTypes := conf.DeviceType // []string{"network", "server", "iot"}

	placeholders := make([]string, len(deviceTypes))
	args := make([]interface{}, len(deviceTypes))

	for i, v := range deviceTypes {
		placeholders[i] = "?"
		args[i] = v
	}

	query := fmt.Sprintf(`
		SELECT *
		FROM devices
		WHERE type IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf("Error query to database: %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)	
		}

		panic(err)
	}
	defer rows.Close()

	// Proses hasil query

	devices := []models.DeviceType{}

	for rows.Next() {
		var d models.DeviceType
		err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.IPAddress,
			&d.Device,
			&d.Error,
			&d.Description,
			&d.DownTime,
			&d.Type,
		)
		if err != nil {
			if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf("Query type mismatch: %v", err)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
			panic(err)
		}
		devices = append(devices, d)
	}

	if err = rows.Err(); err != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf("Error reading rows: %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		panic(err)
	}


	for _, device := range devices {
		wg.Add(1)
		go startDeviceLoop(device,conf, stopChan, &wg)
	}

	wg.Wait()
	fmt.Println("Semua device selesai dimonitor (dihentikan).")
}

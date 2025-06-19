package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"SystemRemoteDevice/template/email"
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"strings"
	"time"
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
	logText := fmt.Sprintf("Time: %s | Name: %s | IP: %s | Device: %s | Lost Percent: %.2f%%\n",
		utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device, lostPercent)

	errWriteTxt := utils.WriteToTXT(outputPath+dev.Name+".txt", logText, true)

	if errWriteTxt != nil {
		fmt.Println("Failed to write file .txt", errWriteTxt)
	} 

	if lostPercent == 100 && !dev.Error {
		setError, errMsg := updateError(dev, true)

		if !setError {
			fmt.Println("Failed to update error status on device:", errMsg)
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
			fmt.Println("Email sent successfully:", message)
		} else {
			fmt.Println("Failed to send email:", message)
		}

	} else if lostPercent < 100 && dev.Error {
		// Kirim email recovery bisa di sini
		setError, errMsg := updateError(dev, false)

		if !setError {
			fmt.Println("Failed to update error status on device:", errMsg)
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
			fmt.Println("Email sent successfully:", message)
		} else {
			fmt.Println("Failed to send email:", message)
		}

	}
}

func MonitoringNetwork(stopChan chan struct{}) {
	fmt.Println("Monitoring network started at",utils.GetCurrentTimeFormatted())

	conf, err := config.LoadJSON[monitoringNetworkType]("config/monitoring-network.json")

	if err != nil {	
		fmt.Println("Failed to load config from json", err)
		return 
	}
	times := conf.Times

	ticker := time.NewTicker(time.Duration(conf.Runtime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			db, err := sql.Open("sqlite", "file:./resource/app.db")
			if err != nil {
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
					panic(err)
				}
				devices = append(devices, d)
			}

			if err = rows.Err(); err != nil {
				panic(err)
			}
			for i := range devices {
				dev := &devices[i]
				replies, err := utils.PingDevice(dev.IPAddress, times)
				if err != nil {
					statusChecking(*dev, 100,conf.OutputPath)
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

				statusChecking(*dev, lostPercent,conf.OutputPath)
			}
		case <-stopChan:
			fmt.Println("Monitoring network stopped at",utils.GetCurrentTimeFormatted())
			return
		}
	}
}

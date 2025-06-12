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

func statusChecking(dev models.DeviceType, lostPercent float64) {
	fmt.Printf("Device Data: %+v\n", dev)
	fmt.Printf("Time: %s | Name: %s | IP: %s | Device: %s | Lost Percent: %.2f%%\n",
		utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device, lostPercent)

	if lostPercent == 100 && !dev.Error {
		setError, errMsg := updateError(dev, true)

		if !setError {
			fmt.Println("Failed to update error status on device:", errMsg)
		} else {
			fmt.Println("Successfully updated error status on device.")
		}

		fmt.Println("Starting email alert process...")

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
		fmt.Println("Email Process Recove")
		setError, errMsg := updateError(dev, false)

		if !setError {
			fmt.Println("Failed to update error status on device:", errMsg)
		} else {
			fmt.Println("Successfully updated error status on device.")
		}

		fmt.Println("Starting email alert process...")

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
	fmt.Println("Monitoring network started")

	times := config.MonitoringNetwork.Times

	ticker := time.NewTicker(time.Duration(config.MonitoringNetwork.Runtime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			db, err := sql.Open("sqlite", "file:./resource/app.db")
			if err != nil {
				panic(err)
			}
			defer db.Close()

			// Query data dengan filter type
			rows, err := db.Query(`
				SELECT *
				FROM devices 
				WHERE type = 'network' OR type = 'server'
			`)
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
				fmt.Printf("ID: %d\nName: %s\nIP: %s\nDevice: %s\nError: %t\nDescription: %s\nDownTime: %v\nType: %s\n\n",
					d.ID, d.Name, d.IPAddress, d.Device, d.Error, d.Description, d.DownTime, d.Type)

				devices = append(devices, d)
			}

			if err = rows.Err(); err != nil {
				panic(err)
			}
			for i := range devices {
				dev := &devices[i]
				replies, err := utils.PingDivice(dev.IPAddress, times)
				if err != nil {
					statusChecking(*dev, 100)
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

				statusChecking(*dev, lostPercent)
			}
		case <-stopChan:
			fmt.Println("Monitoring network stopped")
			return
		}
	}
}

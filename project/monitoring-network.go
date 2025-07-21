package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/template/email"
	"SystemRemoteDevice/utils"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type monitoringNetworkType struct {
	Times      int      `json:"times"`
	Runtime    int      `json:"runtime"`
	DeviceType []string `json:"deviceType"`
	LogPath    string   `json:"logPath"`
	OutputPath string   `json:"outputPath"`
}

func updateError(dev models.DeviceType, errorStatus bool) (bool, string) {
	currentTime := time.Now()

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		return false, fmt.Sprintf("Database open error: %v", err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`UPDATE devices SET error = ?, down_time = ? WHERE id = ? `)
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
func updateCountError(dev models.DeviceType, countError int) (bool, string) {
	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		return false, fmt.Sprintf("Database open error: %v", err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`UPDATE devices SET  count_error = ? WHERE id = ? `)
	if err != nil {
		return false, fmt.Sprintf("Prepare statement error: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(countError, dev.ID)
	if err != nil {
		return false, fmt.Sprintf("Exec update error: %v", err)
	}

	return true, "Success: Error status updated for device"
}

func sendErrorEmail(dev models.DeviceType, conf monitoringNetworkType, linesResArray string) {


	logText := fmt.Sprintf("Time: %s | Name: %s | IP: %s | Device: %s | Result %s",
		utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device, linesResArray)

	errWriteTxt := utils.WriteToTXT(conf.OutputPath+dev.Name+".txt", logText, true)

	if errWriteTxt != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Write Log", fmt.Sprintf("Failed to write file .txt: %v", errWriteTxt)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		return
	}
	setError, errMsg := updateError(dev, true)

	if !setError {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Update Error Status", fmt.Sprintf("Failed to update error status on device: %v", errMsg)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
	}

	email := models.EmailStructure{
		EmailData: models.EmailData{
			Subject:        "Device Down Notification",
			BodyTemplate:   email.ErrorDeviceEmail(dev),
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

func sendRecoveredEmail(dev models.DeviceType, conf monitoringNetworkType, linesResArray string) {
	logText := fmt.Sprintf("Time: %s | Name: %s | IP: %s | Device: %s | Result %s",
		utils.GetCurrentTimeFormatted(), dev.Name, dev.IPAddress, dev.Device, linesResArray)

	errWriteTxt := utils.WriteToTXT(conf.OutputPath+dev.Name+".txt", logText, true)

	if errWriteTxt != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Write Log", fmt.Sprintf("Failed to write file .txt: %v", errWriteTxt)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		return
	}

	setError, errMsg := updateError(dev, false)

	if !setError {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "Update Error Status", fmt.Sprintf("Failed to update error status on device: %v", errMsg)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
	}

	email := models.EmailStructure{
		EmailData: models.EmailData{
			Subject:        "Device Recovery Notification",
			BodyTemplate:   email.RecoveryDeviceEmail(dev),
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

func PingDevice(dev models.DeviceType, conf monitoringNetworkType) {
	times := conf.Times

	replies, errPing := utils.PingDevice(dev.IPAddress, 1)

	lines := strings.Split(replies, "\n")
	linesResArray := lines[2]

	valueARP, err := utils.GetARPEntry(dev.IPAddress)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if !valueARP.Status {
		if !dev.Error {
			sendErrorEmail(dev, conf, linesResArray)
			
			countBol, countMessage := updateCountError(dev, times)
	
			if !countBol {
				if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf("Error query to database: %v", countMessage)); errLog != nil {
					fmt.Printf("Failed to write log: %v\n", errLog)
				}
			}
		}

		return
	}
	if dev.ErrorCount > times {
		dev.ErrorCount = times
	}
	dev.MACAddress = valueARP.MACAddress

	if errPing != nil {
		dev.ErrorCount = utils.Clamp(dev.ErrorCount+1, 0, times)
	} else {
		lines := strings.Split(replies, "\n")
		linesResArray := lines[2]

		if strings.Contains(linesResArray, "Destination Host Unreachable") ||
			strings.Contains(linesResArray, "Request timed out") ||
			strings.Contains(linesResArray, "100% packet loss") ||
			strings.Contains(linesResArray, "Name or service not known") ||
			strings.Contains(linesResArray, "could not find host") {
			dev.ErrorCount = utils.Clamp(dev.ErrorCount+1, 0, times)
		} else {
			dev.ErrorCount = utils.Clamp(dev.ErrorCount-1, 0, times)
		}
	}

	countBol, countMessage := updateCountError(dev, dev.ErrorCount)

	if !countBol {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf("Error query to database: %v", countMessage)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
	}

	if dev.ErrorCount == 0 || dev.ErrorCount == times {
		if errLog := utils.WriteFormattedLog(
			conf.LogPath,
			"INFO",
			"Ping Device",
			fmt.Sprintf("Name: %s | IP: %s | Device: %s | Result %s", dev.Name, dev.IPAddress, dev.Device, linesResArray),
		); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}

		if dev.ErrorCount == times && !dev.Error {
			sendErrorEmail(dev, conf, linesResArray)
		} else if dev.ErrorCount == 0 && dev.Error {
			sendRecoveredEmail(dev, conf, linesResArray)
		}
	}


}

func MonitoringNetwork(stopChan chan struct{}) {
	conf, err := config.LoadJSON[monitoringNetworkType]("config/monitoring-network.json")

	if err != nil {
		fmt.Println("Failed to load config from json", err)
		return
	}

	if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Monitoring Network", fmt.Sprintf("Monitoring network started at %s", utils.GetCurrentTimeFormatted())); errLog != nil {
		fmt.Printf("Failed to write log: %v\n", errLog)
	}

	ticker := time.NewTicker(time.Duration(conf.Times) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			fmt.Println("Monitoring dihentikan oleh sinyal.")
			return
		case <-ticker.C:
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
					&d.ErrorCount,
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
				PingDevice(device, conf)
			}
		}
	}

}

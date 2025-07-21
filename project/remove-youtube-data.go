// package main

package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func verifyYouTubeData(devices []models.DeviceType, outputPath string, logPath string) (bool, string) {
	conf, err := config.LoadJSON[models.AdbConfigType]("config/adb.json")
	if err != nil {
		logToFile(logPath, "ERROR", "Config", "Failed to load config: "+err.Error())
		return false, "Failed to load config: " + err.Error()
	}

	if conf.AdbPath == "" || len(conf.AdbCommandTemplate) == 0 || conf.AdbPort == 0 || len(conf.StatusMessage) == 0 ||
		len(conf.Package) == 0 || conf.VerificationSteps <= 0 || outputPath == "" || len(devices) == 0 {
		logToFile(logPath, "ERROR", "Config", "Incomplete configuration or no devices found.")
		return false, "Incomplete configuration or no devices found"
	}

	startProcessTime := utils.GetCurrentTimeFormatted()

	// Menyimpan status akhir setiap perangkat berdasarkan IP (untuk skip di iterasi selanjutnya)
	statusPerDevice := make(map[string]string)

	// Slice untuk hasil akhir, berurutan sesuai waktu proses
	devicesAfterVerification := []models.DeviceType{}

	for i := 0; i < conf.VerificationSteps; i++ {
		logToFile(logPath, "INFO", "ADB", fmt.Sprintf("Step %d/%d: Restarting ADB server...", i+1, conf.VerificationSteps))

		utils.RunCommand(strings.ReplaceAll(conf.AdbCommandTemplate["kill"], "{adbPath}", conf.AdbPath))
		time.Sleep(5 * time.Second)
		utils.RunCommand(strings.ReplaceAll(conf.AdbCommandTemplate["start"], "{adbPath}", conf.AdbPath))
		time.Sleep(5 * time.Second)

		for j := 0; j < len(devices); j++ {
			ip := devices[j].IPAddress
			deviceName := devices[j].Name

			// Lewati jika sudah sukses pada langkah sebelumnya
			if prevStatus, ok := statusPerDevice[ip]; ok && prevStatus == conf.StatusMessage["SUCCESS"] {
				logToFile(logPath, "INFO", "SKIP", fmt.Sprintf("Skipping %s (%s) - already verified", deviceName, ip))
				continue
			}

			data := map[string]string{
				"adbPath": conf.AdbPath,
				"ip":      ip,
				"port":    strconv.Itoa(conf.AdbPort),
				"package": conf.Package["youtube"],
			}

			connectOutput, err := utils.RunCommand(utils.FillTemplate(conf.AdbCommandTemplate["connect"], data))
			if err != nil || strings.Contains(strings.ToLower(connectOutput), "failed") {
				devices[j].StatusMessage = conf.StatusMessage["FAILED_CONNECT"]
				logToFile(logPath, "ERROR", "CONNECT", fmt.Sprintf("Failed to connect to %s (%s): %s", deviceName, ip, connectOutput))
			} else {
				logToFile(logPath, "INFO", "CLEAR", fmt.Sprintf("Clearing YouTube data on %s (%s)...", deviceName, ip))
				clearOutput, err := utils.RunCommand(utils.FillTemplate(conf.AdbCommandTemplate["clearData"], data))
				if err != nil || strings.Contains(strings.ToLower(clearOutput), "failed") {
					devices[j].StatusMessage = conf.StatusMessage["FAILED_CLEAR"]
					logToFile(logPath, "ERROR", "CLEAR", fmt.Sprintf("Failed to clear data on %s (%s): %s", deviceName, ip, clearOutput))
				} else if strings.Contains(strings.ToLower(clearOutput), "unauthorized") {
					devices[j].StatusMessage = conf.StatusMessage["UNAUTHORIZED"]
					logToFile(logPath, "ERROR", "CLEAR", fmt.Sprintf("Unauthorized access on %s (%s)", deviceName, ip))
				} else {
					devices[j].StatusMessage = conf.StatusMessage["SUCCESS"]
					logToFile(logPath, "INFO", "CLEAR", fmt.Sprintf("Successfully cleared data on %s (%s)", deviceName, ip))
				}
			}

			// Simpan status terbaru ke map
			statusPerDevice[ip] = devices[j].StatusMessage

			// Cek apakah perangkat ini sudah pernah ditambahkan sebelumnya
			found := false
			for k := range devicesAfterVerification {
				if devicesAfterVerification[k].IPAddress == ip {
					devicesAfterVerification[k] = devices[j] // update
					found = true
					break
				}
			}
			if !found {
				devicesAfterVerification = append(devicesAfterVerification, devices[j])
			}
		}
	}
	endProcessTime := utils.GetCurrentTimeFormatted()

	// Format output ke file .txt
	content := fmt.Sprintf(
		"Process Started At : %s\n"+
			"Process Finished At: %s\n\n"+
			"| No | Name Device | IP Address | Status Message\n",
		startProcessTime, endProcessTime,
	)

	for i, device := range devicesAfterVerification {
		line := fmt.Sprintf("| %d | %s | %s | %s\n", i+1, device.Name, device.IPAddress, device.StatusMessage)
		content += line
	}

	if err := utils.WriteToTXT(outputPath, content, false); err != nil {
		logToFile(logPath, "ERROR", "Output", "Failed to write .txt: "+err.Error())
		return false, "Failed to write verification results to log file: " + err.Error()
	}

	logToFile(logPath, "INFO", "Verification", "YouTube data verification completed.")
	return true, "YouTube data verification successful."
}



type removalYoutube	 struct {
	Schedule string `json:"schedule"`
	LogPath  string `json:"logPath"`
	OutputPath  string `json:"outputPath"`
	DeviceType []string `json:"deviceType"`
}

func RemoveYouTubeData() {
// func main() {	
	fmt.Println("YouTube data removal process initiated.")

	conf, errLoadJson := config.LoadJSON[removalYoutube]("config/remove-youtube-data.json")
	if errLoadJson != nil {	
		fmt.Println("Failed to load config from json", errLoadJson)
		return 
	} 
	if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Remove YouTube Data", "Function has been started"); errLog != nil {
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

			deviceTypes := conf.DeviceType 	

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

	status, msg := verifyYouTubeData(devices, conf.OutputPath, conf.LogPath)

	if !status {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "YouTube Data Verification", msg); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		return
	}
	
		email := models.EmailStructure{
		EmailData: models.EmailData{
		Subject:       "YouTube Data Clearance Report - Successful & Failed Devices",
		BodyTemplate:  `
Dear {userName},

Attached is the latest report on the YouTube data clearance process for your network devices.
The report includes details of devices that successfully cleared data and those that encountered errors.

Please review the attached log file for more information.

Best regards, 
Courtyard by Marriott Bali Nusa Dua Resort
			`,
			FileAttachment: []string{
				conf.OutputPath,
			},
		},
	}

	success, message := utils.SendEmail(email)

	if success {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "email", fmt.Sprintf("Email sent successfully: %s", message)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
	} else {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "email", fmt.Sprintf("Failed to send email: %s", message)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		return
	}

	if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Remove YouTube Data", "Function has been completed"); errLog != nil {
		fmt.Printf("Failed to write log: %v\n", errLog)
		return
	}
}
package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"strconv"
	"strings"
	"time"
)

func logToFile(logPath, level, module, message string) {
	_ = utils.WriteFormattedLog(logPath, level, module, message)
}

func getUptime(devices []models.DeviceType, outputPath string, logPath string) (bool, string) {
	conf, err := config.LoadJSON[models.AdbConfigType]("config/adb.json")
	if err != nil {
		logToFile(logPath, "ERROR", "config", "Failed to load config from JSON: "+err.Error())
		return false, "Failed to load config from JSON: " + err.Error()
	}

	if conf.AdbPath == "" || len(devices) == 0 || len(conf.AdbCommandTemplate) == 0 || conf.AdbPort == 0 ||
		len(conf.StatusMessage) == 0 || len(conf.Package) == 0 || conf.VerificationSteps <= 0 || outputPath == "" {
		return false, "Configuration is incomplete."
	}

	devicesMap := make(map[string]models.DeviceType)
	startProcessTime := utils.GetCurrentTimeFormatted()

	for i := 0; i < conf.VerificationSteps; i++ {
		logToFile(logPath, "INFO", "ADB", fmt.Sprintf("Step %d/%d: Restarting ADB server...", i+1, conf.VerificationSteps))
		utils.RunCommand(strings.ReplaceAll(conf.AdbCommandTemplate["kill"], "{adbPath}", conf.AdbPath))
		time.Sleep(5 * time.Second)
		utils.RunCommand(strings.ReplaceAll(conf.AdbCommandTemplate["start"], "{adbPath}", conf.AdbPath))
		time.Sleep(5 * time.Second)

		for j := 0; j < 2; j++ {
			ip := devices[j].IPAddress

			// Skip if already success on previous step
			if i >= 1 {
				if dev, exists := devicesMap[ip]; exists && strings.Contains(dev.StatusMessage,  conf.StatusMessage["SUCCESS"]) {
					logToFile(logPath, "INFO", "ADB", fmt.Sprintf("Skipping %s (%s): Already verified", devices[j].Name, ip))
					continue
				}
			}

			data := map[string]string{
				"adbPath": conf.AdbPath,
				"ip":      ip,
				"port":    strconv.Itoa(conf.AdbPort),
			}

			connectOutput, err := utils.RunCommand(utils.FillTemplate(conf.AdbCommandTemplate["connect"], data))
			if err != nil || strings.Contains(strings.ToLower(connectOutput), "failed") {
				devices[j].StatusMessage = conf.StatusMessage["FAILED_CONNECT"]
				logToFile(logPath, "ERROR", "ADB", fmt.Sprintf("Connection failed to %s (%s): %s", devices[j].Name, ip, connectOutput))
				devicesMap[ip] = devices[j]
				continue
			}

			uptimeOutput, err := utils.RunCommand(utils.FillTemplate(conf.AdbCommandTemplate["getUptime"], data))
			if err != nil || strings.Contains(strings.ToLower(uptimeOutput), "failed") {
				devices[j].StatusMessage = conf.StatusMessage["FAILED_CLEAR"]
				logToFile(logPath, "ERROR", "ADB", fmt.Sprintf("Failed to get uptime for %s", devices[j].Name))
			} else if strings.Contains(strings.ToLower(uptimeOutput), "unauthorized") {
				devices[j].StatusMessage = conf.StatusMessage["UNAUTHORIZED"]
				logToFile(logPath, "ERROR", "ADB", fmt.Sprintf("Unauthorized access to %s", devices[j].Name))
			} else {
				parts := strings.Split(uptimeOutput, " ")
				if len(parts) > 0 {
					uptimeSeconds, err := strconv.ParseFloat(parts[0], 64)
					if err != nil {
						devices[j].StatusMessage = "Failed to parse uptime output"
						logToFile(logPath, "ERROR", "ADB", fmt.Sprintf("Error parsing uptime for %s: %v", devices[j].Name, err))
					} else {
						uptimeDays := uptimeSeconds / (60 * 60 * 24)
						devices[j].StatusMessage = fmt.Sprintf("Success - Uptime %.2f Days", uptimeDays)
						logToFile(logPath, "INFO", "ADB", fmt.Sprintf("Uptime for %s: %.2f days", devices[j].Name, uptimeDays))
					}
				} else {
					devices[j].StatusMessage = "Failed to parse uptime output"
					logToFile(logPath, "ERROR", "ADB", fmt.Sprintf("No output received for uptime: %s", devices[j].Name))
				}
			}

			devicesMap[ip] = devices[j] // Save/update device
		}
	}

	endProcessTime := utils.GetCurrentTimeFormatted()

	// Convert map to slice
	devicesAfterVerification := make([]models.DeviceType, 0, len(devicesMap))
	for _, device := range devicesMap {
		devicesAfterVerification = append(devicesAfterVerification, device)
	}

	// Write to file
	content := fmt.Sprintf("Process Started At : %s\nProcess Finished At: %s\n\n", startProcessTime, endProcessTime)
	content += "| No | Name Device | IP Address | Status Message\n"
	for i, device := range devicesAfterVerification {
		content += fmt.Sprintf("| %d | %s | %s | %s\n", i+1, device.Name, device.IPAddress, device.StatusMessage)
	}

	if err := utils.WriteToTXT(outputPath, content, false); err != nil {
		logToFile(logPath, "ERROR", "ADB", "Failed to write log file: "+err.Error())
		return false, "Failed to write verification results: " + err.Error()
	}

	logToFile(logPath, "INFO", "ADB", "Verification log written successfully.")
	return true, "ADB uptime verification completed."
}

type adbUptime struct {
	Schedule   string   `json:"schedule"`
	LogPath    string   `json:"logPath"`
	OutputPath string   `json:"outputPath"`
	DeviceType []string `json:"deviceType"`
}

func GetUptimeADB() {
	conf, err := config.LoadJSON[adbUptime]("config/get-uptime-ADB-devices.json")
	if err != nil {
		logToFile("logs/default.log", "ERROR", "Config", fmt.Sprintf("Failed to load config JSON: %v", err))
		return
	}

	logToFile(conf.LogPath, "INFO", "Get Uptime ADB Devices", "Function has been started")

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		logToFile(conf.LogPath, "ERROR", "Database", fmt.Sprintf("Error connecting to database: %v", err))
		return
	}
	defer db.Close()

	placeholders := make([]string, len(conf.DeviceType))
	args := make([]interface{}, len(conf.DeviceType))
	for i, v := range conf.DeviceType {
		placeholders[i] = "?"
		args[i] = v
	}

	query := fmt.Sprintf(`SELECT * FROM devices WHERE type IN (%s)`, strings.Join(placeholders, ","))
	rows, err := db.Query(query, args...)
	if err != nil {
		logToFile(conf.LogPath, "ERROR", "Database", fmt.Sprintf("Error executing query: %v", err))
		return
	}
	defer rows.Close()

	devices := []models.DeviceType{}
	for rows.Next() {
		var d models.DeviceType
		err := rows.Scan(&d.ID, &d.Name, &d.IPAddress, &d.Device, &d.Error, &d.Description, &d.DownTime, &d.Type)
		if err != nil {
			logToFile(conf.LogPath, "ERROR", "Database", fmt.Sprintf("Error scanning row: %v", err))
			return
		}
		devices = append(devices, d)
	}

	if err = rows.Err(); err != nil {
		logToFile(conf.LogPath, "ERROR", "Database", fmt.Sprintf("Error finalizing result set: %v", err))
		return
	}

	logToFile(conf.LogPath, "INFO", "Database", fmt.Sprintf("Total devices loaded: %d", len(devices)))

	status, msg := getUptime(devices, conf.OutputPath, conf.LogPath)
	if !status {
		logToFile(conf.LogPath, "ERROR", "Verification", "Error during verification: "+msg)
		return
	}

	email := models.EmailStructure{
		EmailData: models.EmailData{
			Subject: "TV Device Reboot Summary for Guest Rooms",
			BodyTemplate: `
Dear {userName},

Below is a summary of the TV devices in the rooms that require a reboot.
The report includes details of devices that successfully rebooted and those that encountered errors during the process.

Please review the attached log file for more information.

Best regards, 
Courtyard by Marriott Bali Nusa Dua Resort
			`,
			FileAttachment: []string{conf.OutputPath},
		},
	}

	success, message := utils.SendEmail(email)
	if success {
		logToFile(conf.LogPath, "INFO", "Email", "Email sent successfully: "+message)
	} else {
		logToFile(conf.LogPath, "ERROR", "Email", "Failed to send email: "+message)
	}

	logToFile(conf.LogPath, "INFO", "Get Uptime ADB Devices", "Function has been completed")
}

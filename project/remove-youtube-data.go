// package main

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

func verifyYouTubeData(devices []models.DeviceType, outputPath string) (bool, string) {
	conf, err := config.LoadJSON[models.AdbConfigType]("config/adb.json")
	if err != nil {	
		fmt.Println("Failed to load config from json", err)
		return false, "Failed to load config from json: " + err.Error()
	} 

	fmt.Println("Verifying YouTube data...")
	adbPath := conf.AdbPath
	if adbPath == "" {
		return false, "ADB path is not configured."	
	}
	if len(devices) == 0 {
		return false, "No devices found for verification."
	}
	// Ensure the ADB command template is set
	if len(conf.AdbCommandTemplate) == 0 {
		return false, "ADB command template is not configured."
	}
	// Ensure the ADB port is set
	if conf.AdbPort == 0 {
		return false, "ADB port is not configured."
	}
	// Ensure the status messages are set
	if len(conf.StatusMessage) == 0 {
		return false, "Status messages are not configured."
	}
	// Ensure the package names are set
	if len(conf.Package) == 0 {
		return false, "Package names are not configured."
	}
	// Ensure the verification steps are set
	if conf.VerificationSteps <= 0 {
		return false, "Verification steps are not configured."
	}
	// Ensure the log path is set
	if outputPath == "" {	
		return false, "Log path is not configured."
	}

	devicesAfterVerification := []models.DeviceType{}

	startProcessTime := utils.GetCurrentTimeFormatted();

	for i := 0; i < conf.VerificationSteps; i++ {
		utils.RunCommand(strings.ReplaceAll(conf.AdbCommandTemplate["kill"], "{adbPath}", adbPath))
		time.Sleep(5 * time.Second)// Wait for 5 seconds to ensure the ADB server is killed
	
		utils.RunCommand(strings.ReplaceAll(conf.AdbCommandTemplate["start"], "{adbPath}", adbPath))
		time.Sleep(5 * time.Second)// Wait for 5 seconds to ensure the ADB server is started
		
		for j := 0; j < len(devices); i++ {
			if(i >= 1 && devicesAfterVerification[j].StatusMessage == conf.StatusMessage["SUCCESS"]){
				fmt.Println("Skipping verification for device:", devices[j].Name)
				continue
			}

			data := map[string]string{
				"adbPath": adbPath, // ADB path
				"ip":      devices[j].IPAddress, // IP address of the device
				"port":    strconv.Itoa(conf.AdbPort),   // ADB port
				"package": conf.Package["youtube"], // YouTube package name
			}

			// Check if the device is connected
			connectOutput, err := utils.RunCommand(utils.FillTemplate(conf.AdbCommandTemplate["connect"], data))
			if err != nil ||  strings.Contains(strings.ToLower(connectOutput), "failed"){
				devices[j].StatusMessage = conf.StatusMessage["FAILED_CONNECT"]
				devicesAfterVerification = append(devicesAfterVerification, devices[j])
				continue
			}

			clearOutput, err := utils.RunCommand(utils.FillTemplate(conf.AdbCommandTemplate["clearData"], data))
			if err != nil || strings.Contains(strings.ToLower(clearOutput), "failed") {
				devices[j].StatusMessage = conf.StatusMessage["FAILED_CLEAR"]
			} else if strings.Contains(strings.ToLower(clearOutput), "unauthorized") {
				devices[j].StatusMessage = conf.StatusMessage["UNAUTHORIZED"]
			}else {
				devices[j].StatusMessage = conf.StatusMessage["SUCCESS"]
			}

			devicesAfterVerification = append(devicesAfterVerification, devices[j])
		}
	}

	// Write log messages to the file .txt file
	endProcessTime := utils.GetCurrentTimeFormatted();

	content := fmt.Sprintf(
		"Process Started At : %s\n"+
		"Process Finished At: %s\n\n",
		startProcessTime,
		endProcessTime,
	)

	content += "| No | Name Device | IP Address | Status Massage\n"

	// Write the verification results to the log file
	for i , device := range devicesAfterVerification {
		line := fmt.Sprintf("| %d | %s | %s| %s \n", i, device.Name, device.IPAddress , device.StatusMessage)
		content += line
	}

	errWriteTxt := utils.WriteToTXT(outputPath, content, false)
	if errWriteTxt != nil {
		return false, "Failed to write verification results to log file: " + errWriteTxt.Error()
	}

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

	status, msg := verifyYouTubeData(devices, conf.OutputPath)

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
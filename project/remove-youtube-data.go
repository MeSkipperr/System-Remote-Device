// package main

package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func verifyYouTubeData(conf models.AdbConfigType,devices []models.DeviceType,times int, logPath string) (bool, string) {
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
	if logPath == "" {	
		return false, "Log path is not configured."
	}

	devicesAfterVerification := []models.DeviceType{}

	for i := 0; i < times; i++ {
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
	// Write the verification results to the log file
	for _, device := range devicesAfterVerification {
		
	}

	utils.WriteToTXT(logPath, devicesAfterVerification, false)

	return true, "YouTube data verification successful."
}

func RemoveYouTubeData() {
// func main() {	
	fmt.Println("YouTube data removal process initiated.")

	conf, err := config.LoadJSON[models.AdbConfigType]("config/adb.json")
	if err != nil {	
		fmt.Println("Failed to load config from json", err)
		return 
	}
	devices := []models.DeviceType{
		{
			ID:          1,
			Name:        "Device1",	
			Type:        "network",
		},
		{
			ID:          2,
			Name:        "Device1",	
			Type:		 "server",
		},
		{
			ID:          3,
			Name:        "Device1",	
			Type:		 "iot",
		},
	}

	status, msg := verifyYouTubeData(conf, devices, conf.VerificationSteps, "logs/youtube_data_removal.log")

	if !status {
		fmt.Println("Error during YouTube data verification:", msg)
		return
	}
		

	fmt.Println("YouTube data has been removed successfully.")
}
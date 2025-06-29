package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	"database/sql"

	_ "modernc.org/sqlite"
)

type systemHasError	 struct {
	LogPath  string `json:"logPath"`
	OutputPath  string `json:"outputPath"`
	DeviceType []string `json:"deviceType"`
}

func CheckSystemHasError() {
	conf, errLoadJson := config.LoadJSON[systemHasError]("config/check-system-has-error.json")
	if errLoadJson != nil {	
		fmt.Println("Failed to load config from json", errLoadJson)
		return 
	} 
	if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Check System Has Error", "Function has been started"); errLog != nil {
		fmt.Printf("Failed to write log: %v\n", errLog)
	}
	outputPath := conf.OutputPath;

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
		WHERE error = 1 AND type IN (%s)
	`, strings.Join(placeholders, ","))


	rows, err := db.Query(query, args...)
	if err != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf("Error query to database database: %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		panic(err)
	}
	defer rows.Close()

	errorDevices := []models.DeviceType{}

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
		errorDevices = append(errorDevices, d)
	}

	if err = rows.Err(); err != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "database", fmt.Sprintf(" %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		panic(err)
	}

	// Buat file Excel baru
	f := excelize.NewFile()

	// Nama sheet
	sheet := "Error Devices"
	index, errSheet := f.NewSheet(sheet)
	if errSheet != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "xlsx", fmt.Sprintf("Error to make sheet %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		return
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")


	// Header kolom
	f.SetCellValue(sheet, "A1", "No")
	f.SetCellValue(sheet, "B1", "Name")
	f.SetCellValue(sheet, "C1", "Ip Address")
	f.SetCellValue(sheet, "D1", "Device")
	f.SetCellValue(sheet, "E1", "Error")
	f.SetCellValue(sheet, "F1", "Down Time")
	f.SetCellValue(sheet, "G1", "Description")

	
	// Tulis data
	for i, device := range errorDevices {
		formattedTime := device.DownTime.Format("03:04:05 PM - 02/01/2006")

		row := i + 2 // mulai dari baris ke-2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), device.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), device.IPAddress)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), device.Device)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), device.Error)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), formattedTime)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), device.Description)
	}

	// Simpan ke file
	if errSheet := f.SaveAs(outputPath); errSheet != nil {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "xlsx", fmt.Sprintf("Error to save file %v", errSheet)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)		
		}
	} else {
		if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "xlsx", fmt.Sprintf("File successfully created in: %s", outputPath)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
	}

		email := models.EmailStructure{
		EmailData: models.EmailData{
			Subject:       "Notification of Network Device Status - Error Detected",
			BodyTemplate:  `
Dear {userName},

Please find attached a collection of devices that still have errors detected in your system. This report provides detailed information about the affected devices for your review.

Best regards,
Courtyard by Marriott Bali Nusa Dua Resort
			`,
			FileAttachment: []string{
				outputPath,
			},
		},
	}

	success, message := utils.SendEmail(email)

	if success {
		fmt.Println("Email sent successfully:", message)
		if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "email", fmt.Sprintf("Email sent successfully: %s", message)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)		
		}
	} else {
		fmt.Println("Failed to send email:", message)
		if errLog := utils.WriteFormattedLog(conf.LogPath, "ERROR", "email", fmt.Sprintf("Failed to send email: %s", message)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		return
	}

	if errLog := utils.WriteFormattedLog(conf.LogPath, "INFO", "Check System Has Error", "Function has been completed"); errLog != nil {
		fmt.Printf("Failed to write log: %v\n", errLog)
		return
	}
}

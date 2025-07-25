package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"bytes"
	"fmt"
	"strings"

	"database/sql"

	"golang.org/x/crypto/ssh"
	_ "modernc.org/sqlite"
)



type deviceCommandType struct {
	Device     string   `json:"device"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Command    []string `json:"command"`
	SpecificIP []string `json:"specificIP"`
}

type configType struct{
	Device 		[]deviceCommandType		`json:"deviceConfig"`
	LogPath		string					`json:"logPath"`
	OutputPath		string				`json:"outputPath"`
}

// ---------- CORE FUNCTION ----------
func RunSSHCommands(device models.DeviceType,deviceCommand deviceCommandType) (string, error) {
	var outputBuilder strings.Builder

		// SSH configuration
		config := &ssh.ClientConfig{
			User:            deviceCommand.Username,
			Auth:            []ssh.AuthMethod{ssh.Password(deviceCommand.Password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		// Connect to device
		client, err := ssh.Dial("tcp", device.IPAddress+":22", config)
		if err != nil {
			return "", fmt.Errorf("failed to connect to %s via SSH: %v", device.IPAddress, err)
		}
		defer client.Close()

		// Write IP address as header
		outputBuilder.WriteString("=======\n")
		outputBuilder.WriteString("Device Name: " + device.Name + "\n")
		outputBuilder.WriteString("IP Address: " + device.IPAddress + "\n")
		outputBuilder.WriteString("Device Model: " + device.Device + "\n")
		outputBuilder.WriteString("Device Information: " + device.Description + "\n")
		outputBuilder.WriteString("=======\n")

		// Execute all commands
		for _, cmd := range deviceCommand.Command {
			session, err := client.NewSession()
			if err != nil {
				return "", fmt.Errorf("failed to create session for %s: %v", device.IPAddress, err)
			}

			var stdout bytes.Buffer
			session.Stdout = &stdout

			if err := session.Run(cmd); err != nil {
				session.Close()
				return "", fmt.Errorf("failed to run command %q on %s: %v", cmd, device.IPAddress, err)
			}
			session.Close()

			outputBuilder.WriteString("Command: " + cmd + "\n\n")
			outputBuilder.WriteString(stdout.String())
			if !strings.HasSuffix(stdout.String(), "\n") {
				outputBuilder.WriteByte('\n')
			}
			outputBuilder.WriteString("=======\n")
		}

	return outputBuilder.String(), nil
}


func GetSystemInformation() {
	configData, errCommand := config.LoadJSON[configType]("config/command-information.json")
	
	if errCommand != nil{
		fmt.Println("Failed to load config from json", errCommand)
		return 
	}
	deviceCommand := configData.Device
	outputPath := configData.OutputPath
	
	if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Get System Information", "Get System Information Starting..."); errLog != nil {
		fmt.Printf("Failed to write log: %v\n", errLog)
	}
	
	selectedDevices := []string{}

	for _, d := range deviceCommand{
		selectedDevices = append(selectedDevices, d.Device)
	}

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		if errLog := utils.WriteFormattedLog(configData.LogPath, "ERROR", "database", fmt.Sprintf("Error connecting to database: %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		panic(err)
	}
	defer db.Close()

	// Buat placeholder ?, ?, ? sebanyak isi slice
	placeholders := make([]string, len(selectedDevices))
	args := make([]interface{}, len(selectedDevices))
	for i, val := range selectedDevices {
		placeholders[i] = "?"
		args[i] = val
	}

	query := fmt.Sprintf(`
		SELECT * FROM devices
		WHERE device IN (%s)
	`, strings.Join(placeholders, ", "))

	// Eksekusi query dengan argumen slice
	rows, err := db.Query(query, args...)

	if err != nil {
		if errLog := utils.WriteFormattedLog(configData.LogPath, "ERROR", "database", fmt.Sprintf("Error query to database: %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)	
		}
		panic(err)
	}
	defer rows.Close()

	
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
			if errLog := utils.WriteFormattedLog(configData.LogPath, "ERROR", "database", fmt.Sprintf("Query type mismatch: %v", err)); errLog != nil {
				fmt.Printf("Failed to write log: %v\n", errLog)
			}
			panic(err)
		}

		devices = append(devices, d)
	}

	if err = rows.Err(); err != nil {
		if errLog := utils.WriteFormattedLog(configData.LogPath, "ERROR", "database", fmt.Sprintf(" %v", err)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		panic(err)
	}


	var summaryOut strings.Builder
	for _, device := range devices {
		var commandList deviceCommandType
		found := false

		outerLoop:
		for _, cmd := range deviceCommand {
			if cmd.Device != device.Device {
				continue
			}

			// Prioritaskan yang memiliki SpecificIP
			if len(cmd.SpecificIP) > 0 {
				for _, ip := range cmd.SpecificIP {
					if ip == device.IPAddress {
						commandList = cmd
						found = true
						break outerLoop
					}
				}
			} else {
				// Simpan sementara jika belum ketemu yang SpecificIP
				if !found {
					commandList = cmd
				}
			}
		}
		if len(commandList.Command) > 0 && commandList.Device != ""{
			out, err := RunSSHCommands(device, commandList)
			if err != nil {
				fmt.Println(err)
				summaryOut.WriteString("=======\n")
				summaryOut.WriteString("Device Name: " + device.Name + "\n")
				summaryOut.WriteString("IP Address: " + device.IPAddress + "\n")
				summaryOut.WriteString("Device Model: " + device.Device + "\n")
				summaryOut.WriteString("Device Information: " + device.Description + "\n")
				summaryOut.WriteString("=======\n")
				summaryOut.WriteString(err.Error() + "\n")
				summaryOut.WriteString("=======\n")
			}
			summaryOut.WriteString(out+"\n")
		}
	}

	errWriteTxt := utils.WriteToTXT(outputPath, summaryOut.String(),false)
	if errWriteTxt != nil {
		fmt.Println("Error:", errWriteTxt)
		if errLog := utils.WriteFormattedLog(configData.LogPath, "ERROR", "Get System Information", fmt.Sprintf("Failed to write data in file: %v", errWriteTxt)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)	
		}
		return
	} else {
		fmt.Println("Success write data in file : ",outputPath)
		if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Get System Information", fmt.Sprintf("Success write data in file: %s", outputPath)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)	
		}
	}

	email := models.EmailStructure{
		EmailData: models.EmailData{
			Subject:       "Device Information",
			BodyTemplate:  `
Dear {userName},
The attached report contains detailed information about the current status of devices in the network. This information has been generated to help identify and address potential issues within the system.

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
		if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Email", fmt.Sprintf("Email sent successfully: %s", message)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
	} else {
		if errLog := utils.WriteFormattedLog(configData.LogPath, "ERROR", "Email", fmt.Sprintf("Failed to send email: %s", message)); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
		}
		return
	}

	if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Get System Information", "Function has been completed"); errLog != nil {
		fmt.Printf("Failed to write log: %v\n", errLog)
	}
}

package project

import (
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"SystemRemoteDevice/utils/network"
	"fmt"
	"github.com/robfig/cron/v3"
)

type NetworkInterface struct {
	InterfaceName string `json:"interfaceName"`
	Name          string `json:"name"`
}
type ScheduleContentType struct {
	Cron    string `json:"cron"`
	Enabled bool   `json:"enabled"`
}
type ScheduleSType struct {
	GetSpeedtestNetwork  ScheduleContentType `json:"getSpeedtestNetwork"`
	SendSpeedtestNetwork ScheduleContentType `json:"sendSpeedtestNetwork"`
}
type ConfigSpeedTest struct {
	Network    []NetworkInterface `json:"network"`
	Schedule   ScheduleSType      `json:"schedule"`
	LogPath    string             `json:"logPath"`
	OutputPath string             `json:"outputPath"`
}

	func getResult() {
		configData, errCommand := config.LoadJSON[ConfigSpeedTest]("config/speedtest.json")

		if errCommand != nil {
			fmt.Println("Failed to load config from json", errCommand)
			return
		}
		if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Get Speed Test Network", "Get Speed Test Network Starting..."); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
			return
		}
		networkDetails := network.GetIPAddress()
		if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Get Speed Test Network", "Get Speed Test Network Success"); errLog != nil {
			fmt.Printf("Failed to write log: %v\n", errLog)
			return
		}

		for _, detail := range networkDetails {
			for _, netCfg := range configData.Network {
				if netCfg.InterfaceName != detail.InterfaceName {
					continue
				}

				speedResult, err := network.RunningSpeedtest(detail.IPAddress)
				if err != nil {
					fmt.Printf("Failed to run speed test: %v\n", err)
					utils.WriteFormattedLog(configData.LogPath, "ERROR", "SpeedTest", fmt.Sprintf("Failed to run speed test: %v", err))
					continue
				}

				resultText := fmt.Sprintf(
					"Time        : %s\n"+
						"Network     : %s\n"+
						"Interface   : %s\n"+
						"   IP Address : %s\n"+
						"   Gateway    : %s\n"+
						"   Netmask    : %s\n"+
						"ISP        : %s\n"+
						"IP Public  : %s\n"+
						"Ping       : %d ms\n"+
						"Server Name: %s\n"+
						"Country    : %s\n"+
						"Download   : %.2f Mbps\n"+
						"Upload     : %.2f Mbps\n"+
						"--------------------------------------------------------\n",
					utils.GetCurrentTimeFormatted(),
					netCfg.Name,
					detail.InterfaceName,
					detail.IPAddress,
					detail.Gateway,
					detail.Netmask,
					speedResult.ISP,
					speedResult.PublicIP,
					speedResult.PingMs,
					speedResult.ServerName,
					speedResult.Country,
					speedResult.Download,
					speedResult.Upload,
				)

				if(speedResult.Download == 0 && speedResult.Upload == 0) {
					resultText = fmt.Sprintf("No internet connection detected on %s", netCfg.Name)
				}

				if(speedResult.Download < 10.0 || speedResult.Upload < 10.0) {
					body := fmt.Sprintf(`
Dear {userName},

Internet speed is currently low/unstable (one or both directions are below 10 Mbps).
Below is the latest measurement for your reference:

Time       : %s
Interface  : %s
Network    : %s
IP Address : %s
Gateway    : %s
Netmask    : %s
ISP        : %s
Public IP  : %s
Ping       : %d ms
Download   : %.2f Mbps
Upload     : %.2f Mbps

Impact: You may experience slow browsing, buffering, and call interruptions.
We are monitoring and will inform you once performance returns to normal.

Best regards,
Courtyard by Marriott Bali Nusa Dua Resort`,
    utils.GetCurrentTimeFormatted(),
    detail.InterfaceName,
    netCfg.Name,
    detail.IPAddress,
    detail.Gateway,
    detail.Netmask,
    speedResult.ISP,
    speedResult.PublicIP,
    speedResult.PingMs,
    speedResult.Download,
    speedResult.Upload,
)

					email := models.EmailStructure{
					EmailData: models.EmailData{
					Subject: "Speed Test Network Results",
					BodyTemplate: body,
							FileAttachment: []string{
								configData.OutputPath,
							},
						},
					}

					success, message := utils.SendEmail(email)

					if success {
						utils.WriteFormattedLog(configData.LogPath, "INFO", "email", fmt.Sprintf("Email sent successfully: %s", message))
					} else {
						utils.WriteFormattedLog(configData.LogPath, "ERROR", "email", fmt.Sprintf("Failed to send email: %s", message))
					}
				}

				errWriteTxt := utils.WriteToTXT(configData.OutputPath, resultText, true)
				if errWriteTxt != nil {
					fmt.Println("Error:", errWriteTxt)
					utils.WriteFormattedLog(configData.LogPath, "ERROR", "WriteTXT", fmt.Sprintf("Failed to write data: %v", errWriteTxt))
				}
			}
		}

	}

func GetSpeedTestNetwork() {
    configData, err := config.LoadJSON[ConfigSpeedTest]("config/speedtest.json")
    if err != nil {
        fmt.Println("Failed to load config:", err)
        return
    }

    c := cron.New()

    // Schedule speedtest network retrieval
    if configData.Schedule.GetSpeedtestNetwork.Enabled {
        cronExpr := configData.Schedule.GetSpeedtestNetwork.Cron
        if cronExpr == "" {
            cronExpr = "0 */2 * * *" // default: tiap 2 jam
        }

        _, err := c.AddFunc(cronExpr, func() {
            if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Speedtest", "Starting test..."); errLog != nil {
                fmt.Printf("Failed to write log: %v\n", errLog)
                return
            }

            getResult()

            if errLog := utils.WriteFormattedLog(configData.LogPath, "INFO", "Speedtest", "Test completed successfully"); errLog != nil {
                fmt.Printf("Failed to write log: %v\n", errLog)
            }
        })
        if err != nil {
            fmt.Println("Failed to schedule GetSpeedtestNetwork:", err)
        }
    }

    // Schedule sending email with results
    if configData.Schedule.SendSpeedtestNetwork.Enabled {
        cronExpr := configData.Schedule.SendSpeedtestNetwork.Cron
        if cronExpr == "" {
            cronExpr = "0 10 * * *" // default: jam 10 pagi
        }

        _, err := c.AddFunc(cronExpr, func() {
            email := models.EmailStructure{
                EmailData: models.EmailData{
                    Subject: "Speed Test Network Results",
                    BodyTemplate: `
Dear {userName},

Attached is the daily summary of speed test network results collected over the past 24 hours.  
The detailed report can be found in the attached .txt file.

Best regards, 
Courtyard by Marriott Bali Nusa Dua Resort
                    `,
                    FileAttachment: []string{
                        configData.OutputPath,
                    },
                },
            }

            success, message := utils.SendEmail(email)

            if success {
                utils.WriteFormattedLog(configData.LogPath, "INFO", "email", fmt.Sprintf("Email sent successfully: %s", message))
            } else {
                utils.WriteFormattedLog(configData.LogPath, "ERROR", "email", fmt.Sprintf("Failed to send email: %s", message))
            }

            // kosongkan file output setelah terkirim
            if err := utils.WriteToTXT(configData.OutputPath, "", false); err != nil {
                fmt.Println("Error:", err)
                utils.WriteFormattedLog(configData.LogPath, "ERROR", "WriteTXT", fmt.Sprintf("Failed to clear output: %v", err))
            }
        })
        if err != nil {
            fmt.Println("Failed to schedule SendSpeedtestNetwork:", err)
        }
    }

    // start cron scheduler
    c.Start()
}

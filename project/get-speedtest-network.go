package project

import (
	"fmt"
	"SystemRemoteDevice/config"
	"SystemRemoteDevice/utils"
	"SystemRemoteDevice/utils/network"
)
type NetworkInterface struct {
	InterfaceName string `json:"interfaceName"`
	Name       string `json:"name"`
}

type ServerID struct {
	ID int `json:"id"`	
	Name string `json:"name"`
}
type ConfigSpeedTest struct {
	Network []NetworkInterface `json:"network"`
	ServerID []ServerID `json:"serverId"`
	LogPath    string   `json:"logPath"`
	OutputPath string   `json:"outputPath"`

}
func GetSpeedTestNetwork (){

	configData, errCommand := config.LoadJSON[ConfigSpeedTest]("config/speedtest.json")
	
	if errCommand != nil{
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
    if detail.IsDHCP {
        for _, netCfg := range configData.Network {
            if netCfg.InterfaceName == detail.InterfaceName {
                for _, server := range configData.ServerID {
                    // Convert server.ID dari string ke int

                    // Run SpeedTest
                    speedResult, err := network.RunningSpeedtest(network.SpeedTestParms{
                        SourceIP: detail.IPAddress,
                        ServerID: server.ID,
                    })
                    if err != nil {
                        fmt.Printf("Failed to run speed test: %v\n", err)
                        utils.WriteFormattedLog(configData.LogPath, "ERROR", "SpeedTest", fmt.Sprintf("Failed to run speed test: %v", err))
                        continue
                    }

                    // Format hasil untuk ditulis ke file
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
                        speedResult.Timestamp,
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

                    // Append ke TXT file
                    errWriteTxt := utils.WriteToTXT(configData.OutputPath, resultText, true)
                    if errWriteTxt != nil {
                        fmt.Println("Error:", errWriteTxt)
                        utils.WriteFormattedLog(configData.LogPath, "ERROR", "WriteTXT", fmt.Sprintf("Failed to write data: %v", errWriteTxt))
                    } else {
                        fmt.Println("Success write data in file :", configData.OutputPath)
                        utils.WriteFormattedLog(configData.LogPath, "INFO", "WriteTXT", fmt.Sprintf("Success write data in file: %s", configData.OutputPath))
                    }
                }
            }
        }
    }
}

}
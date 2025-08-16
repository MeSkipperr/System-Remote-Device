package network

import (
	"fmt"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

type SpeedResult struct {
	SourceIP   string  `json:"source_ip"`
	PublicIP   string  `json:"public_ip"`
	ISP        string  `json:"isp"`
	PingMs     int64   `json:"ping_ms"`
	Download   float64 `json:"download_mbps"`
	Upload     float64 `json:"upload_mbps"`
	Timestamp  string  `json:"timestamp"`
	Country    string  `json:"country"`
	ServerID   string `xml:"id,attr" json:"id"`
	ServerName string  `json:"name"`
}

type SpeedTestParms struct {
	SourceIP string `json:"source_ip"`
	ServerID int    `json:"server_id"`
}

func RunningSpeedtest(data SpeedTestParms) (SpeedResult, error) {
	// Validate SourceIP
	if data.SourceIP == "" {
		return SpeedResult{}, fmt.Errorf("SourceIP cannot be empty")
	}

	client := speedtest.New()
	speedtest.WithUserConfig(&speedtest.UserConfig{Source: data.SourceIP})(client)

	user, err := client.FetchUserInfo()
	if err != nil {
		return SpeedResult{}, fmt.Errorf("failed to fetch user info: %v", err)
	}

	serverList, err := client.FetchServers()
	if err != nil {
		return SpeedResult{}, fmt.Errorf("failed to fetch server list: %v", err)
	}

	var targets speedtest.Servers
	if data.ServerID == 0 {
		// Use nearest server if ServerID is not provided
		targets, _ = serverList.FindServer([]int{})
	} else {
		// Use specified ServerID
		targets, _ = serverList.FindServer([]int{data.ServerID})
	}

	if len(targets) == 0 {
		return SpeedResult{}, fmt.Errorf("no server found")
	}

	s := targets[0]
	s.PingTest(nil)
	s.DownloadTest()
	s.UploadTest()

	result := SpeedResult{
		SourceIP:   data.SourceIP,
		PublicIP:   user.IP,
		ISP:        user.Isp,
		Country:    s.Country,
		ServerID:   s.ID,
		ServerName: s.Name,
		PingMs:     s.Latency.Milliseconds(),
		Download:   s.DLSpeed.Mbps(),
		Upload:     s.ULSpeed.Mbps(),
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

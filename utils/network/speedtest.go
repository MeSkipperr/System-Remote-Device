package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

type cliResult struct {
	ISP       string `json:"isp"`
	Interface struct {
		ExternalIP string `json:"externalIp"`
	} `json:"interface"`
	Ping struct {
		Latency float64 `json:"latency"`
	} `json:"ping"`
	Download struct {
		Bandwidth float64 `json:"bandwidth"`
	} `json:"download"`
	Upload struct {
		Bandwidth float64 `json:"bandwidth"`
	} `json:"upload"`
	Server struct {
		ID      json.Number `json:"id"`
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"server"`
}

// struct hasil kamu tetap sama
type SpeedResult struct {
	SourceIP   string  `json:"source_ip"`
	PublicIP   string  `json:"public_ip"`
	ISP        string  `json:"isp"`
	PingMs     int64   `json:"ping_ms"`
	Download   float64 `json:"download_mbps"`
	Upload     float64 `json:"upload_mbps"`
	Timestamp  string  `json:"timestamp"`
	Country    string  `json:"country"`
	ServerID   string  `json:"id"`
	ServerName string  `json:"name"`
}

func RunningSpeedtest(sourceIp string) (SpeedResult, error) {
	if sourceIp == "" {
		return SpeedResult{}, fmt.Errorf("SourceIP cannot be empty")
	}

	// path relatif ke project
	exePath := filepath.Join(".", "resource", "speedtest.exe")

	// jalankan speedtest CLI dengan output JSON
	cmd := exec.Command(exePath, "--accept-license", "--accept-gdpr", "-f", "json")
	out, err := cmd.Output()
	if err != nil {
		return SpeedResult{}, fmt.Errorf("failed to run speedtest CLI: %w", err)
	}

	// parse JSON ke struct sementara
	var cli cliResult
	if err := json.Unmarshal(out, &cli); err != nil {
		return SpeedResult{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// bandwidth dari bps → Mbps
	downloadMbps := cli.Download.Bandwidth / 125000.0
	uploadMbps := cli.Upload.Bandwidth / 125000.0

	result := SpeedResult{
		SourceIP:   sourceIp,
		PublicIP:   cli.Interface.ExternalIP,
		ISP:        cli.ISP,
		Country:    cli.Server.Country,
		ServerID:   cli.Server.ID.String()	,
		ServerName: cli.Server.Name,
		PingMs:     int64(cli.Ping.Latency),
		Download:   downloadMbps,
		Upload:     uploadMbps,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

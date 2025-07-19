package utils

import (
	"os/exec"
	"strings"
)

type ARPEntry struct {
	IPAddress   string
	MACAddress  string
	Type        string
	Status      bool
}

func GetARPEntry(ip string) (ARPEntry, error) {
	cmd := exec.Command("arp", "-a", ip)
	outputBytes, err := cmd.Output()
	if err != nil {
		return ARPEntry{
			IPAddress: ip,
			Status:    false,
		}, nil
	}

	output := string(outputBytes)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ip) {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return ARPEntry{
					IPAddress:  fields[0],
					MACAddress: fields[1],
					Type:       fields[2],
					Status:     true,
				}, nil
			}
		}
	}

	return ARPEntry{
		IPAddress: ip,
		Status:    false,
	}, nil
}

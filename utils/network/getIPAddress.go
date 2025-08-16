package network

import (
	"net"
	"os/exec"
	"regexp"
	"strings"
)

type NetworkDetails struct {
	InterfaceName string
	IPAddress     string
	Netmask       string
	Gateway       string
	DNS           []string
	IsDHCP        bool
}

func GetIPAddress() []NetworkDetails {
	cmd := exec.Command("netsh", "interface", "ip", "show", "config")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	results := parseNetshOutput(string(output))

	return results 
}

func parseNetshOutput(output string) []NetworkDetails {
	// Pisah berdasarkan "Configuration for interface"
	re := regexp.MustCompile(`(?m)^Configuration for interface .+`)
	matches := re.FindAllStringIndex(output, -1)

	var results []NetworkDetails
	for i := 0; i < len(matches); i++ {
		start := matches[i][0]
		end := len(output)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		block := output[start:end]
		lines := strings.Split(strings.TrimSpace(block), "\n")

		var details NetworkDetails
		for _, line := range lines {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "Configuration for interface") {
				details.InterfaceName = strings.Trim(line[len("Configuration for interface")+1:], `"`)
			}
			if strings.Contains(line, "DHCP enabled") {
				details.IsDHCP = strings.Contains(strings.ToLower(line), "yes")
			}
			if strings.Contains(line, "IP Address") {
				details.IPAddress = extractLastWord(line)
			}
			if strings.Contains(line, "Subnet Prefix") {
				details.Netmask = extractWithRegex(line, `mask ([\d\.]+)`)
			}
			if strings.Contains(line, "Default Gateway") {
				gw := extractLastWord(line)
				if gw != "None" {
					details.Gateway = gw
				}
			}
			if strings.Contains(line, "DNS Servers") || strings.Contains(line, "Statically Configured DNS Servers") {
				dns := extractLastWord(line)
				if net.ParseIP(dns) != nil {
					details.DNS = append(details.DNS, dns)
				}
			}
		}
		results = append(results, details)
	}
	return results
}


func extractLastWord(s string) string {
	parts := strings.Fields(s)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func extractWithRegex(s, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

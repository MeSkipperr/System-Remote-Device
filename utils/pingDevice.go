package utils

import (
	"fmt"
	"os/exec"
	"strconv"
)

func PingDevice(ip string, times int) (string, error) {
	if ip == "" {
		return "", fmt.Errorf("IP address cannot be empty")
	}
	if 0 >= times {
		times = 1
	}
	var cmd *exec.Cmd
	countStr := strconv.Itoa(times)
	cmd = exec.Command("ping", "-n", countStr, ip)

	out, err := cmd.CombinedOutput()

	return string(out), err
}
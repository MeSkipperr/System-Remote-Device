// runCommand.go
// utils/runCommand.go
package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunCommand(command string) (string, error) {
	args := strings.Fields(command)

	if len(args) == 0 {
		return "", fmt.Errorf("command is empty")
	}

	cmd := exec.Command(args[0], args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command error: %v\nOutput: %s", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}
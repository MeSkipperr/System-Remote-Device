package schedule
// utils/schedule/restart.go
import (
	"fmt"
	"os/exec"
	"runtime"
)
func RestartComputer() {
	var command *exec.Cmd

	if runtime.GOOS == "windows" {
		command = exec.Command("shutdown", "/r", "/t", "0")
	} else {
		command = exec.Command("sudo", "reboot")
	}

	output, err := command.CombinedOutput()
	if err != nil {
		fmt.Println("Failed to execute restart command:", err)
		return
	}

	fmt.Println("Restart command executed successfully:", string(output))
}
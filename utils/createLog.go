package utils

import (
	"fmt"
	"os"
	"time"
)

func WriteFormattedLog(path, level, module, message string) error {
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s\n",
		time.Now().Format("03:04:05 PM - 02/01/2006"),
		level,
		module,
		message,
	)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(logLine)
	return err
}

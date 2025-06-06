package utils

import "time"

func GetCurrentTimeFormatted() string {
	now := time.Now()
	return now.Format("03:04:05 PM - 02/01/2006")
}
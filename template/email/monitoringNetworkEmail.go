package email

import (
	"SystemRemoteDevice/models"
	"SystemRemoteDevice/utils"
	"strings"
	"fmt"
)

func ErrorDeviceEmail(dev models.DeviceType) string{
	currentTime := utils.GetCurrentTimeFormatted()
	
	desc := ""
	if trimmed := strings.TrimSpace(dev.Description); trimmed != "" {
		desc = fmt.Sprintf("- Descriptions : %s\n", trimmed)
	}

return fmt.Sprintf(`
Dear {userName}

We would like to inform you that an error has occurred in the network system. Below are the details:

- Time : %s
- Host Name: %s
- IP Address: %s
- Device: %s
%s
Kindly review the information provided and take necessary actions to resolve the issue at your earliest convenience.

Best regards,
Courtyard by Marriott Bali Nusa Dua Resort
`, currentTime, dev.Name, dev.IPAddress, dev.Device, desc)
}

func RecoveryDeviceEmail(dev models.DeviceType) string{
	currentTime := utils.GetCurrentTimeFormatted()
	
	desc := ""
	if trimmed := strings.TrimSpace(dev.Description); trimmed != "" {
		desc = fmt.Sprintf("- Descriptions : %s\n", trimmed)
	}

return fmt.Sprintf(`
Dear {userName}

This is to notify you of a network system recovery update. Below are the recovery details:

- Time : %s
- Host Name: %s
- IP Address: %s
- Device: %s
%s
Inspect the details and confirm the system is back to normal.

Best regards,
Courtyard by Marriott Bali Nusa Dua Resort
`, currentTime, dev.Name, dev.IPAddress, dev.Device, desc)
}


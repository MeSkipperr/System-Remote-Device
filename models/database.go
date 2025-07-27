package models

import "time"

type DeviceType struct {
	ID            int       `json:"id"`
	Name          string    `json:"name" validate:"required"`
	IPAddress     string    `json:"ip_address" validate:"required"`
	Device        string    `json:"device" validate:"required"`
	Error         bool      `json:"error"`
	Description   string    `json:"description"`
	DownTime      time.Time `json:"down_time" validate:"required"`
	Type          string    `json:"type" validate:"required"`
	StatusMessage string    `json:"status_message"`
	ErrorCount    int       `json:"count_error"`
	MACAddress    string    `json:"mac_address"`
}
type UserType struct{
	ID	   int    `json:"user_id"`       // Unique identifier for the user
	Email    string `json:"user_email"`         // User's email address
	Password string `json:"user_password"`      // User's password
	Role     string `json:"user_role"`          // User's role (e.g., admin
	Name	 string `json:"user_name"`          // User's name
}
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
	Name	 string `json:"user_name"`          // User's name
	Email    string `json:"user_email"`         // User's email address
	Role     string `json:"user_role"`          // User's role (e.g., admin
	Password string `json:"user_pass"`      // User's password
}

type VerificationToken struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Code      string    `json:"code" db:"code"` 
	Type      string    `json:"type" db:"type"` 
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

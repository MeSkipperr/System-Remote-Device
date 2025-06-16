package models

import "time"

type DeviceType struct {
    ID          int      
    Name        string    
    IPAddress   string    
    Device      string    
    Error       bool      
    Description string    
    DownTime    time.Time
    Type        string    
    StatusMessage  string // Optional field to store error messages
}

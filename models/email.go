package models

// User represents an individual email sender or recipient.
type Recipients struct {
	UserName string `json:"userName"` // Username
	Email     string `json:"email"`     // User's email address
}
type Sender struct {
	Email     string `json:"email"`     // SMTP email
	Pass      string `json:"pass"`      // SMTP password
	Name      string `json:"name"`      // SMTP Name 
}


// EmailData contains the subject, body, and attachments of the email message.
type EmailData struct {
	Subject        string   `json:"subject"`        // Subject line of the email
	BodyTemplate   string   `json:"bodyTemplate"`   // Body content or template of the email
	FileAttachment []string `json:"fileAttachment"` // List of file attachment URLs or paths
}

// EmailStructure defines the complete structure required to send an email.
type EmailStructure struct {
	Recipients []Recipients     `json:"recipients"` // List of users receiving the email
	EmailData  EmailData  `json:"emailData"`  // Email content details
	Sender 	   Sender     `json:"sender"` // List of users sending the email
}


package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"database/sql"
	_ "modernc.org/sqlite"

	"github.com/joho/godotenv"

	"SystemRemoteDevice/config"
	"SystemRemoteDevice/models"
)

type smtpType struct {
	SMTPHost string		`json:"smtpHost"`
	SMTPPort string		`json:"smtpPort"`
	Name     string		`json:"name"`
}

// SendEmail delivers an e‑mail to every recipient listed in the provided
// EmailStructure.  It supports two modes:
//  1. Plain‑text only (when FileAttachment is empty)
//  2. Multipart/MIME with one or more attachments (when FileAttachment has paths)
//
// The function performs extensive validation on the incoming data, falls back to
// .env credentials when none are supplied, and returns a boolean indicating
// overall success plus an aggregated message describing any errors that
// occurred for individual recipients.
//
// Parameters
// ----------
// email : models.EmailStructure
//
//	The aggregate payload containing sender credentials, recipient list,
//	message subject/body template, optional attachments, and contextual
//	device data.
//
// Returns
// -------
// ok : bool
//
//	true  → every message sent successfully
//	false → at least one validation or send error occurred
//
// msg : string
//
//	A human‑readable success string or a concatenated list of problems.
func SendEmail(email models.EmailStructure) (bool, string) {
	//------------------------------------------------
	// 0. ── Local variable shortcuts
	//------------------------------------------------
	recipients := email.Recipients
	emailData := email.EmailData
	sender := email.Sender
	var errors []string // collects validation + send failures

	//------------------------------------------------
	// 1. ── Mandatory field validation
	//------------------------------------------------
	if strings.TrimSpace(emailData.Subject) == "" {
		errors = append(errors, "email subject cannot be empty")
	}
	if strings.TrimSpace(emailData.BodyTemplate) == "" {
		errors = append(errors, "email bodyTemplate cannot be empty")
	}

	// Pull SMTP credentials from environment if they were not supplied
	if strings.TrimSpace(sender.Email) == "" || strings.TrimSpace(sender.Pass) == "" {
		_ = godotenv.Load() // ignore error – env vars might already exist
		if sender.Email == "" {
			sender.Email = os.Getenv("EMAIL_USER")
		}
		if sender.Pass == "" {
			sender.Pass = os.Getenv("EMAIL_PASS")
		}
		if sender.Email == "" || sender.Pass == "" {
			errors = append(errors, "SMTP credentials cannot be empty")
		}
	}

	// Verify recipient slice is not empty and that each field is filled in
	if len(recipients) == 0 {

		db, err := sql.Open("sqlite", "file:./resource/app.db")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		rows, err := db.Query(`
            SELECT user_name , user_email  
            FROM users 
            WHERE user_role = 'recipient' OR user_role = 'support'
        `)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		user := []models.Recipients{}

		for rows.Next() {
			var u models.Recipients
			err := rows.Scan(
                &u.UserName,
                &u.Email,
			)
			if err != nil {
				panic(err)
			}

            user = append(user, u)
		}

		recipients = user
	}

	for i, r := range recipients {
		if strings.TrimSpace(r.UserName) == "" {
			errors = append(errors, fmt.Sprintf("UserName at index %d cannot be empty", i))
		}
		if strings.TrimSpace(r.Email) == "" {
			errors = append(errors, fmt.Sprintf("Email at index %d cannot be empty", i))
		}
	}

	//Get data config from json
	smtpConfig, errSMTP := config.LoadJSON[smtpType]("config/smtp.json")
	
	if errSMTP != nil {
		errors = append(errors, fmt.Sprintf("Failed to load SMTP config: %v", errSMTP))
	}
	if len(errors) > 0 {
		return false, strings.Join(errors, "; ")
	}

	// Early exit on validation failure
	errors = errors[:0] // reset for sending phase

	//------------------------------------------------
	// 2. ── Set up SMTP authentication (PLAIN)
	//------------------------------------------------
	auth := smtp.PlainAuth("", sender.Email, sender.Pass, smtpConfig.SMTPHost)

	//------------------------------------------------
	// 3. ── Send loop per recipient (ensures personalised body)
	//------------------------------------------------
	for _, user := range recipients {
		// 3a. Personalise body placeholders
		body := strings.ReplaceAll(emailData.BodyTemplate, "{userName}", user.UserName)

		// 3b. Build message headers + payload
		fromHeader := fmt.Sprintf("%s <%s>",smtpConfig.Name, sender.Email)
		toHeader := user.Email
		var msg bytes.Buffer

		if len(emailData.FileAttachment) == 0 {
			// ---- Plain‑text path ----
			msg.WriteString("Subject: " + emailData.Subject + "\r\n" +
				"From: " + fromHeader + "\r\n" +
				"To: " + toHeader + "\r\n" +
				"\r\n" +
				body)
		} else {
			// ---- Multipart/MIME with attachments ----
			boundary := "BOUNDARY-Email-Go-123456"
			msg.WriteString("Subject: " + emailData.Subject + "\r\n" +
				"From: " + fromHeader + "\r\n" +
				"To: " + toHeader + "\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/mixed; boundary=" + boundary + "\r\n" +
				"\r\n")

			// Part 1: text body
			msg.WriteString("--" + boundary + "\r\n")
			msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			msg.WriteString(body + "\r\n")

			// Part 2..n: each attachment path
			for _, path := range emailData.FileAttachment {
				data, err := os.ReadFile(path)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s: failed to read %s (%v)", user.Email, path, err))
					continue // skip this attachment but still attempt sending
				}

				filename := filepath.Base(path)
				mimeType := mime.TypeByExtension(filepath.Ext(filename))
				if mimeType == "" {
					mimeType = "application/octet-stream"
				}

				msg.WriteString("--" + boundary + "\r\n")
				msg.WriteString("Content-Type: " + mimeType + "\r\n")
				msg.WriteString("Content-Disposition: attachment; filename=\"" + filename + "\"\r\n")
				msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

				// Encode and wrap at 76 chars per RFC 2045
				enc := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
				base64.StdEncoding.Encode(enc, data)
				for i := 0; i < len(enc); i += 76 {
					end := i + 76
					if end > len(enc) {
						end = len(enc)
					}
					msg.Write(enc[i:end])
					msg.WriteString("\r\n")
				}
			}

			// Closing boundary
			msg.WriteString("--" + boundary + "--")
		}

		// 3c. Attempt to send
		if err := smtp.SendMail(
			smtpConfig.SMTPHost+":"+smtpConfig.SMTPPort,
			auth,
			sender.Email,         // envelope‑from
			[]string{user.Email}, // envelope‑to
			msg.Bytes(),
		); err != nil {
			errors = append(errors, fmt.Sprintf("Cannot send to email: %s. Error: %v", user.Email, err))
		}
	}

	//------------------------------------------------
	// 4. ── Summarise outcome
	//------------------------------------------------
	if len(errors) > 0 {
		return false, strings.Join(errors, "; ")
	}
	return true, "Success to send email to all users"
}

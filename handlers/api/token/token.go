package token

import (
	"net/http"
	"SystemRemoteDevice/models"
	"encoding/json"
	_ "modernc.org/sqlite"
	"database/sql"
	"SystemRemoteDevice/utils"
	"time"
	"SystemRemoteDevice/handlers/api"
	"SystemRemoteDevice/internal/security"

)


func CreateToken(w http.ResponseWriter, r *http.Request) {
	var params models.VerificationToken 
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if params.Email == ""   || params.Type == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	expiresAt := time.Now().Add(5 * time.Minute)

	generateCode,err :=  utils.GenerateVerificationCode(6) // Generate a random token string

	if(err != nil) {
		http.Error(w, "Error generating verification code", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	query := `
	INSERT INTO verification_tokens (email, code, type, expires_at)
	VALUES (?, ?, ?, ?)`

	_ , err = db.Exec(query,
		params.Email,
		generateCode,
		"register",
		expiresAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	email := models.EmailStructure{
		EmailData: models.EmailData{
			Subject:        "Verification Code",
			BodyTemplate:   "Your Verfication Code is:" + generateCode, 
			FileAttachment: []string{},
		},
		Recipients: []models.Recipients{
			{
				UserName: params.Email,
				Email:    params.Email,
			},
		},
	}

	success ,message := utils.SendEmail(email)

	if !success {
		http.Error(w, "Failed to send email: "+message, http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Token generated successfully",
	})
}

func VerifyToken(w http.ResponseWriter, r *http.Request) {
	var token models.VerificationToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if token.Email == "" || token.Code == "" || token.Type == "" {
		http.Error(w, "Email and code are required", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Langkah 1: Verifikasi token
	query := `
	SELECT * FROM verification_tokens 
	WHERE email = ? AND code = ? AND type = ? AND expires_at > datetime('now') AND used = 0`
	rows, err := db.Query(query, token.Email, token.Code, token.Type)
	if err != nil {
		http.Error(w, "Failed to verify token", http.StatusInternalServerError)
		return
	}

	found := rows.Next()
	rows.Close() 

	if !found {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	query = `UPDATE verification_tokens
	SET used = 1
	WHERE email = ? AND code = ? AND type = ?`
	_, err = db.Exec(query, token.Email, token.Code, token.Type)
	if err != nil {
		http.Error(w, "Failed to update token status", http.StatusInternalServerError)
		return
	}

	validToken, err := security.GenerateToken(security.TokenParams{
		Payload: map[string]interface{}{
			"user_email": token.Email,
			"token_code": token.Code,
		},
		ExpiresIn: "30m", 
	})

	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Token verified successfully",
		"token":   validToken,
	})
}
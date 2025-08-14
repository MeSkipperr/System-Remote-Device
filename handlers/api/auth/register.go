package auth

import (
	"SystemRemoteDevice/handlers/api"
	"SystemRemoteDevice/internal/security"
	"SystemRemoteDevice/models"
	"database/sql"
	"encoding/json"
	"net/http"

	_ "modernc.org/sqlite"
)

type UserParams struct {
	models.UserType
	Token   string `json:"token"`             // Authentication token
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	var user UserParams
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if user.Name == "" || user.Email == "" || user.Password == "" || user.Role == "" || user.Token == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := security.BcryptHash(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	

	query := "INSERT INTO users (user_name, user_email, user_role, user_pass) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(query, user.Name, user.Email, user.Role, hashedPassword)

	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusCreated, map[string]string{"status": "User added successfully"})
}

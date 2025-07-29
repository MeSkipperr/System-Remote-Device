package api

import (
	"SystemRemoteDevice/internal/security"
	"SystemRemoteDevice/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

func GetUserQuery(query string, arg interface{}) ([]models.UserType, error) {
db, err := sql.Open("sqlite", "file:./resource/app.db")
if err != nil {
	return nil, err
}
defer db.Close()

rows, err := db.Query(query, arg)
if err != nil {
	return nil, err
}
defer rows.Close()

var users []models.UserType
for rows.Next() {
	var d models.UserType
	err := rows.Scan(
		&d.ID,
		&d.Name,
		&d.Email,
		&d.Role,
		&d.Password,
	)
	if err != nil {
		return nil, err
	}
	users = append(users, d)
}

if err := rows.Err(); err != nil {
	return nil, err
}

return users, nil	 
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	query := "SELECT * FROM users"
	users, err := GetUserQuery(query, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userWithoutPassword := make([]models.UserType, 0, len(users))
	for _, user := range users {
		userWithoutPassword = append(userWithoutPassword, models.UserType{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		})
	}
	WriteJSON(w, http.StatusOK, userWithoutPassword)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	var user models.UserType
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if user.Name == "" || user.Email == "" || user.Password == "" || user.Role == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	fmt.Printf("Adding user: %s\n", user.Name)
	fmt.Printf("User email: %s\n", user.Email)
	fmt.Printf("User role: %s\n", user.Role)
	fmt.Printf("User password: %s\n", user.Password)

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

	WriteJSON(w, http.StatusCreated, map[string]string{"status": "User added successfully"})
}

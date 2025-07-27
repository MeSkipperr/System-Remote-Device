package api 

import (
	"SystemRemoteDevice/models"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
	"log"
)

 func getUserQuery(query string, arg interface{}) ([]models.UserUserType, error) {
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
			&d.Password,
			&d.Role,
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
	users, err := getUserQuery(query, nil)
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
	writeJSON(w, http.StatusOK, userWithoutPassword)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	var user models.UserType
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(query, user.Name, user.Email, user.Password, user.Role)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "User added successfully"})
}
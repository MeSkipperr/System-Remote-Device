package api

import (
	"SystemRemoteDevice/models"
	"database/sql"
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


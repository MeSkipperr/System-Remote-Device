package auth

import (
	"SystemRemoteDevice/handlers/api"
	"SystemRemoteDevice/internal/security"
	"SystemRemoteDevice/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Implement login logic here
	// This will typically involve checking the user's credentials against a database
	// and returning a session token or error response.

	var loginData models.UserType
	if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if loginData.Name == "" || loginData.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	
	query := "SELECT * FROM users WHERE user_name = ? "
	users, err := api.GetUserQuery(query, loginData.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("All users: %+v\n", users[0])
	if len(users) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if ( !security.CheckBcryptHash(loginData.Password, users[0].Password)) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	token, err := security.GenerateToken(security.TokenParams{
		Payload: map[string]interface{}{
			"user_email":  users[0].Email,
			"user_name":  users[0].Name,
			"user_id":  users[0].ID,
		},
		ExpiresIn: "1h", // 30 days
	})
	if( err != nil) {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",               
		HttpOnly: true,
		Secure:   true,              
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * time.Minute), // Sesuaikan dengan token expires
	})
	api.WriteJSON(w, http.StatusOK, map[string]string{
		"token": token,
		"message": "Login successful",
	})
	
	// If login is successful, you might want to set a session cookie or return a token
}

// fetch("https://yourdomain.com/api/protected", {
//   method: "GET", // atau POST, PUT, dll.
//   credentials: "include" // ⬅️ wajib agar cookie dikirim
// })
// .then(res => res.json())
// .then(data => {
//   console.log(data)
// });


// cara mendapatkan token di frontend dan bisa di gunakan untuk request ke endpoint yang membutuhkan otentikasi``
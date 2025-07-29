package security

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Ambil secret dari env atau fallback
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

// Claims berisi data yang disimpan dalam token
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
type TokenParams struct {
	Payload   map[string]interface{}
	ExpiresIn string // contoh: "1h", "7d"
}

func GenerateToken(params TokenParams) (string, error) {
	// Map durasi
	durationMap := map[string]time.Duration{
		"1m":  time.Minute,
		"30m": 30 * time.Minute,
		"1h":  time.Hour,
		"6h":  6 * time.Hour,
		"12h": 12 * time.Hour,
		"1d":  24 * time.Hour,
		"7d":  7 * 24 * time.Hour,
		"30d": 30 * 24 * time.Hour,
	}

	duration, ok := durationMap[params.ExpiresIn]
	if !ok {
		duration = time.Hour
	}
	expirationTime := time.Now().Add(duration)

	// Isi claims
	claims := jwt.MapClaims{}
	for key, val := range params.Payload {
		claims[key] = val
	}
	claims["exp"] = expirationTime.Unix()

	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}


// VerifyToken memverifikasi token JWT dan mengembalikan claims-nya
func VerifyToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}

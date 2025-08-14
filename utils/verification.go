package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// GenerateVerificationCode generates a numeric verification code with the specified length.
// Example: GenerateVerificationCode(6) => "825691"
func GenerateVerificationCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid code length")
	}

	var sb strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10)) // 0-9
		if err != nil {
			return "", err
		}
		sb.WriteString(n.String())
	}

	return sb.String(), nil
}
	
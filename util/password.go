package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword return bcrypt hashed password using the default cost(10)
func HashPassword(password string) (string, error) {
	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("unable to hash password %v", err)
	}

	return string(bs), nil

}

// CheckPassword checks if plainPassword matches hashedPassword
func CheckPassword(plainPassword, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

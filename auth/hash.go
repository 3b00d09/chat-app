package auth

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func GeneratHashedPassword(password string) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Fatal(err)
	}

	return hashedPassword
}

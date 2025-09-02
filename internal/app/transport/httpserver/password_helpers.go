package httpserver

import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) (string, error) {
	bitesPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bitesPassword), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

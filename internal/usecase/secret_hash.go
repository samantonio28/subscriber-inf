package usecase

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashSecret хэширует пароль алгоритмом bcrypt. В БД хранится ТОЛЬКО хэш —
// требование ЛР №5: пароли не должны лежать в явном виде.
func HashSecret(secret string) (string, error) {
	if len(secret) < 6 {
		return "", fmt.Errorf("secret is too short (minimum 6 characters)")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash secret: %w", err)
	}
	return string(hash), nil
}

// CheckSecret сравнивает открытый пароль с bcrypt-хэшем.
func CheckSecret(hash, secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret)) == nil
}

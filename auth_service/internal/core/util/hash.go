package util

import "golang.org/x/crypto/bcrypt"

type hash struct{}

var Hash hash

// HashearPassword convierte una contraseña en un hash seguro con bcrypt
func (h hash) HashearPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

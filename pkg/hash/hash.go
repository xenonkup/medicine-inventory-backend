package hash

import "golang.org/x/crypto/bcrypt"

// Password hashes a plaintext password using bcrypt.
func Password(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Check reports whether plain matches the bcrypt hash.
func Check(plain, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

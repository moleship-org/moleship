package port

// PasswordManager defines the interface for password hashing and comparison.
type PasswordManager interface {
	Hash(password string) (string, error)
	Compare(passwordHash, password string) (bool, error)
}

// TokenGenerator defines the interface for generating tokens and their corresponding hashes.
type TokenGenerator interface {
	Generate() (token string, hash []byte, err error)
}

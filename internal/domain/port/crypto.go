package port

type PasswordManager interface {
	Hash(password string) (string, error)
	Compare(passwordHash, password string) (bool, error)
}

type TokenGenerator interface {
	Generate() (token string, hash []byte, err error)
	GenerateWithExpiry(minutes int) (token string, hash []byte, err error)
}

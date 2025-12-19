package auth

// Service defines the authentication business logic
type Service interface {
	Login(username, password string) (string, error)
	ValidateToken(token string) (Claims, error)
}

type Claims struct {
	UserID string
	Role   string
}

package auth

// Service defines the authentication business logic
type Service interface {
	Login(username, password string) (string, error)
	ValidateToken(token string) (Claims, error)
}

type Claims struct {
	UserID   string
	Role     string
	Identity Identity
}

// MockProvider implements simple token-to-role mapping with identity enrichment
type MockProvider struct {
	resolver IdentityResolver
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		resolver: NewMockIdentityResolver(),
	}
}

func (p *MockProvider) Login(username, password string) (string, error) {
	// Mock successful login for any user
	return "mock-token-" + username, nil
}

func (p *MockProvider) ValidateToken(token string) (Claims, error) {
	identity, err := p.resolver.Resolve(nil, token)
	if err != nil {
		return Claims{}, nil // Invalid token
	}

	return Claims{
		UserID:   identity.UID,
		Role:     "analyst", // Default to analyst for mock unless it's admin
		Identity: identity,
	}, nil
}

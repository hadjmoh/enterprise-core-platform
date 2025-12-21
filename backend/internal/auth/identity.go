package auth

import (
	"context"
	"fmt"
	"strings"
)

// Identity represents a unified organizational actor
type Identity struct {
	UID         string   `json:"uid"`
	PrimaryUser string   `json:"primary_user"`
	Email       string   `json:"email"`
	Department  string   `json:"department"`
	Title       string   `json:"title"`
	Status      string   `json:"status"`
	ManagedIPs  []string `json:"managed_ips"`
}

// IdentityResolver converts various identifiers into a unified Identity
type IdentityResolver interface {
	Resolve(ctx context.Context, identifier string) (Identity, error)
}

// MockIdentityResolver provides static mapping for development
type MockIdentityResolver struct {
	store map[string]Identity
}

func NewMockIdentityResolver() *MockIdentityResolver {
	store := map[string]Identity{
		"token-admin": {
			UID:         "UID-001",
			PrimaryUser: "admin_superuser",
			Email:       "admin@enterprise.net",
			Department:  "Infrastructure",
			Title:       "Platform Architect",
			Status:      "active",
		},
		"token-analyst": {
			UID:         "UID-002",
			PrimaryUser: "analyst_jdoe",
			Email:       "jdoe@enterprise.net",
			Department:  "SOC-Blue",
			Title:       "Security Analyst L2",
			Status:      "active",
			ManagedIPs:  []string{"10.0.0.50", "192.168.1.100"},
		},
		"10.0.0.50": {
			UID:         "UID-002",
			PrimaryUser: "analyst_jdoe",
			Department:  "SOC-Blue",
		},
	}

	return &MockIdentityResolver{store: store}
}

func (r *MockIdentityResolver) Resolve(ctx context.Context, identifier string) (Identity, error) {
	// Normalize identifier (e.g., lower-case for usernames)
	id := strings.ToLower(identifier)
	
	if identity, ok := r.store[id]; ok {
		return identity, nil
	}

	// Token-based fallback for existing mock tokens
	if strings.HasPrefix(id, "token-") {
		if identity, ok := r.store[id]; ok {
			return identity, nil
		}
	}

	return Identity{}, fmt.Errorf("identity not found for identifier: %s", identifier)
}

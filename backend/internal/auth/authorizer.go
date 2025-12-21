package auth

import (
	"strings"
)

// Permission levels for commands
var rolePermissions = map[string]map[string]bool{
	"admin": {
		"search":    true,
		"where":     true,
		"fields":    true,
		"eval":      true,
		"limit":     true,
		"stats":     true,
		"bucket":    true,
		"timechart": true,
		"join":      true, // Admin only
		"soar":      true, // Admin only
	},
	"analyst": {
		"search":    true,
		"where":     true,
		"fields":    true,
		"eval":      true,
		"limit":     true,
		"stats":     true,
		"bucket":    true,
		"timechart": true,
	},
	"auditor": {
		"search": true,
	},
}

// Authorizer manages RBAC for SPL commands
type Authorizer struct{}

func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

// IsCommandAllowed checks if a role has permission to run a specific command
func (a *Authorizer) IsCommandAllowed(role string, command string) bool {
	p, ok := rolePermissions[strings.ToLower(role)]
	if !ok {
		return false
	}
	return p[strings.ToLower(command)]
}

// GetDataScope returns the allowed index range for a role
func (a *Authorizer) GetDataScope(role string) string {
	switch strings.ToLower(role) {
	case "auditor":
		return "_audit"
	case "analyst":
		return "*" // For now, analyst can see everything except high-priv indices (to be refined)
	case "admin":
		return "*"
	default:
		return "none"
	}
}

package enrichment

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
)

// UserEnricher adds user context to events
type UserEnricher struct {
	users  map[string]*UserInfo
	logger *logger.Logger
	mu     sync.RWMutex
}

type UserInfo struct {
	FullName   string
	Email      string
	Department string
	Role       string
	Manager    string
	Title      string
	Location   string
}

func NewUserEnricher(logger *logger.Logger) *UserEnricher {
	return &UserEnricher{
		users:  make(map[string]*UserInfo),
		logger: logger,
	}
}

func (u *UserEnricher) Enrich(data map[string]interface{}) map[string]interface{} {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// Try to find user by username or email
	userFields := []string{"user", "username", "email", "principal"}
	
	var userKey string
	for _, field := range userFields {
		if val, ok := data[field].(string); ok {
			userKey = val
			break
		}
	}

	if userKey == "" {
		return data
	}

	if user, ok := u.users[userKey]; ok {
		data["user_full_name"] = user.FullName
		data["user_email"] = user.Email
		data["user_department"] = user.Department
		data["user_role"] = user.Role
		data["user_manager"] = user.Manager
		data["user_title"] = user.Title
		data["user_location"] = user.Location
	}

	return data
}

func (u *UserEnricher) LoadUsers(users map[string]*UserInfo) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.users = users
	u.logger.Info("Loaded users", "count", len(users))
}

func (u *UserEnricher) AddUser(key string, info *UserInfo) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.users[key] = info
}

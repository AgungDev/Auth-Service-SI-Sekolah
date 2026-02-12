package dto

// CreateUserRequest represents create user input
type CreateUserRequest struct {
	Email    string
	Password string
	TenantID string
	RoleIDs  []string
}

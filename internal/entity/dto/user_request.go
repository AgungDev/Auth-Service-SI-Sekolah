package dto

// CreateUserRequest represents create user input
type CreateUserRequestBody struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	TenantID string   `json:"tenant_id"`
	RoleIDs  []string `json:"role_ids"`
}
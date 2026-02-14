package dto

// CreateUserRequest represents create user input
type CreateUserRequestBody struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	TenantID string   `json:"tenant_id"`
	RoleIDs  []string `json:"role_ids"`
}

// UpdateUserRequestBody represents fields allowed to be updated for a user
type UpdateUserRequestBody struct {
	Email    string   `json:"email,omitempty"`
	Password string   `json:"password,omitempty"`
	RoleIDs  []string `json:"role_ids,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`
}
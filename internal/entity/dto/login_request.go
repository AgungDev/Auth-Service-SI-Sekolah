package dto

// LoginRequest represents login input
type LoginRequest struct {
	Email    string
	Password string
	TenantID string
}

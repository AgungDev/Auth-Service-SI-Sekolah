package dto

// LoginRequest represents login input
// type LoginRequest struct {
// 	Email    string
// 	Password string
// 	TenantID string
// }

type LoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id"`
}
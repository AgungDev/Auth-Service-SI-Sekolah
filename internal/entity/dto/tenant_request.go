package dto

// CreateTenantRequest represents create tenant input
type CreateTenantRequestBody struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Status  string `json:"status,omitempty"` // optional, default ACTIVE
}
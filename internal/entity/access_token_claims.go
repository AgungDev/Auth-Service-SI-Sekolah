package entity

type AccessTokenClaims struct {
	Sub          string   `json:"sub"`       // user-id
	TenantID     string   `json:"tenant_id"` // school-id
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	TenantStatus string   `json:"tenant_status"`
	ExpiresAt    int64    `json:"exp"`
}

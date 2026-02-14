package entity

type AccessTokenClaims struct {
	Sub          string   `json:"sub"`           // user_uuid
	TenantID     string   `json:"tenant_id"`     // tenant_uuid or empty for SUPER_ADMIN
	Role         string   `json:"role"`          // single role
	Permissions  []string `json:"permissions"`   // permissions array
	IsSuperAdmin bool     `json:"is_super_admin"`
	ExpiresAt    int64    `json:"exp"`
	IssuedAt     int64    `json:"iat"`
	Issuer       string   `json:"iss"`
}

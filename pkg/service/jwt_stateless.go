package service

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// StatelessClaims is a lightweight claims model usable by other services
// without importing internal packages.
type StatelessClaims struct {
    Sub          string   `json:"sub"`
    TenantID     string   `json:"tenant_id"`
    Role         string   `json:"role"`
    Permissions  []string `json:"permissions"`
    IsSuperAdmin bool     `json:"is_super_admin"`
    ExpiresAt    int64    `json:"exp"`
    IssuedAt     int64    `json:"iat"`
    Issuer       string   `json:"iss"`
}

// JWTStatelessValidator validates JWT tokens without any DB calls.
// It is safe for other services to import from pkg/service and use.
type JWTStatelessValidator struct {
    signingKey    []byte
    signingMethod jwt.SigningMethod
    expectedIss   string
}

// NewStatelessValidator creates a new validator.
func NewStatelessValidator(signingKey []byte, method jwt.SigningMethod, expectedIssuer string) *JWTStatelessValidator {
    return &JWTStatelessValidator{
        signingKey:    signingKey,
        signingMethod: method,
        expectedIss:   expectedIssuer,
    }
}

// Verify parses and validates a token string and returns StatelessClaims.
// Accepts token with or without "Bearer " prefix.
func (v *JWTStatelessValidator) Verify(tokenString string) (*StatelessClaims, error) {
    if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
        tokenString = strings.TrimSpace(tokenString[7:])
    }

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // verify signing method
        if v.signingMethod != nil {
            if token.Method.Alg() != v.signingMethod.Alg() {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
        }
        return v.signingKey, nil
    })

    if err != nil {
        return nil, fmt.Errorf("token parse error: %w", err)
    }

    if !token.Valid {
        return nil, ErrInvalidToken
    }

    claimsMap, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, ErrInvalidToken
    }

    // optional: validate issuer
    if v.expectedIss != "" {
        if iss, _ := claimsMap["iss"].(string); iss != v.expectedIss {
            return nil, fmt.Errorf("invalid issuer")
        }
    }

    // extract permissions
    perms := []string{}
    if pi, ok := claimsMap["permissions"].([]interface{}); ok {
        for _, v := range pi {
            if s, ok := v.(string); ok {
                perms = append(perms, s)
            }
        }
    }

    sub, _ := claimsMap["sub"].(string)
    tenantID, _ := claimsMap["tenant_id"].(string)
    roleStr, _ := claimsMap["role"].(string)
    isSuper, _ := claimsMap["is_super_admin"].(bool)
    expF, _ := claimsMap["exp"].(float64)
    iatF, _ := claimsMap["iat"].(float64)
    iss, _ := claimsMap["iss"].(string)

    return &StatelessClaims{
        Sub:          sub,
        TenantID:     tenantID,
        Role:         roleStr,
        Permissions:  perms,
        IsSuperAdmin: isSuper,
        ExpiresAt:    int64(expF),
        IssuedAt:     int64(iatF),
        Issuer:       iss,
    }, nil
}

// HasPermission helper to check permission in claims
func (c *StatelessClaims) HasPermission(code string) bool {
    for _, p := range c.Permissions {
        if p == code {
            return true
        }
    }
    return false
}

package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"auth-service/internal/usecase"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	jwtService  usecase.JWTService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUseCase *usecase.AuthUseCase, jwtService usecase.JWTService) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		jwtService:  jwtService,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// LoginRequest represents the login request body
type LoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id"`
}

// Login handles POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req LoginRequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	ctx := r.Context()
	resp, err := h.authUseCase.Login(ctx, usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		TenantID: req.TenantID,
	})

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// RefreshTokenRequestBody represents the refresh token request body
type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken handles POST /refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req RefreshTokenRequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	ctx := r.Context()
	resp, err := h.authUseCase.RefreshToken(ctx, usecase.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// CreateTenantRequestBody represents the create tenant request body
type CreateTenantRequestBody struct {
	Name string `json:"name"`
}

// CreateTenant handles POST /tenants
func (h *AuthHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Check authorization - SUPER_ADMIN only
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing authorization token"})
		return
	}

	claims, err := h.jwtService.VerifyAccessToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid token"})
		return
	}

	// Check if user has SUPER_ADMIN role
	isAdmin := false
	for _, role := range claims.Roles {
		if role == "SUPER_ADMIN" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Insufficient permissions"})
		return
	}

	var req CreateTenantRequestBody
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	ctx := r.Context()
	tenant, err := h.authUseCase.CreateTenant(ctx, usecase.CreateTenantRequest{
		Name: req.Name,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Tenant created successfully",
		Data:    tenant,
	})
}

// CreateUserRequestBody represents the create user request body
type CreateUserRequestBody struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	TenantID string   `json:"tenant_id"`
	RoleIDs  []string `json:"role_ids"`
}

// CreateUser handles POST /users
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Check authorization
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing authorization token"})
		return
	}

	claims, err := h.jwtService.VerifyAccessToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid token"})
		return
	}

	var req CreateUserRequestBody
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Check if user has permission to create users in this tenant
	if claims.TenantID != req.TenantID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Cannot create users in other tenants"})
		return
	}

	ctx := r.Context()
	user, err := h.authUseCase.CreateUser(ctx, usecase.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		TenantID: req.TenantID,
		RoleIDs:  req.RoleIDs,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "User created successfully",
		Data:    user,
	})
}

// Health handles GET /health
func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Auth Service is running",
		Data:    nil,
	})
}

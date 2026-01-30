package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"auth-service/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userModel *models.UserModel
	jwtSecret []byte
}

func NewAuthHandler(userModel *models.UserModel, secret string) *AuthHandler {
	return &AuthHandler{
		userModel: userModel,
		jwtSecret: []byte(secret),
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Register creates a new user account
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		h.sendError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	if len(req.Password) < 8 {
		h.sendError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Create user
	ctx := context.Background()
	user, err := h.userModel.CreateUser(ctx, req.Email, string(hashedPassword), req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			h.sendError(w, http.StatusConflict, "email already exists")
			return
		}
		h.sendError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	h.sendJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "user created successfully",
		"user_id": user.ID,
		"email":   user.Email,
	})
}

// Login authenticates a user and returns JWT tokens
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		h.sendError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Get user by email
	ctx := context.Background()
	user, err := h.userModel.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.sendError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.sendError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Generate tokens
	accessToken, err := h.generateAccessToken(user.ID, user.Email)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	refreshToken, err := h.generateRefreshToken(user.ID, user.Email)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}

	// Store refresh token in Redis
	if err := h.userModel.StoreRefreshToken(ctx, user.ID, refreshToken); err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to store refresh token")
		return
	}

	h.sendJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
	})
}

// RefreshToken generates a new access token using a refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Parse refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		h.sendError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		h.sendError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	userID := int64(claims["user_id"].(float64))
	email := claims["email"].(string)

	// Verify refresh token exists in Redis
	ctx := context.Background()
	if !h.userModel.ValidateRefreshToken(ctx, userID, req.RefreshToken) {
		h.sendError(w, http.StatusUnauthorized, "refresh token not found or expired")
		return
	}

	// Generate new access token
	accessToken, err := h.generateAccessToken(userID, email)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"accessToken": accessToken,
		"expiresIn":   900,
	})
}

// Logout invalidates the refresh token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Parse token to get user ID
	token, _ := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		userID := int64(claims["user_id"].(float64))
		ctx := context.Background()
		h.userModel.RevokeRefreshToken(ctx, userID)
	}

	h.sendJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// VerifyToken validates a JWT token
func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.sendError(w, http.StatusUnauthorized, "missing authorization header")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.sendError(w, http.StatusUnauthorized, "invalid authorization header format")
		return
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		h.sendError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		h.sendJSON(w, http.StatusOK, map[string]interface{}{
			"valid":   true,
			"user_id": claims["user_id"],
			"email":   claims["email"],
		})
	} else {
		h.sendError(w, http.StatusUnauthorized, "invalid token claims")
	}
}

func (h *AuthHandler) generateAccessToken(userID int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) generateRefreshToken(userID int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, map[string]string{"error": message})
}

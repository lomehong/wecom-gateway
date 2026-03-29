package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"wecom-gateway/internal/store"
)

const (
	// TokenExpiration is the default JWT token expiration time (24 hours)
	TokenExpiration = 24 * time.Hour
)

var (
	// ErrInvalidCredentials is returned when username or password is invalid
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrUserDisabled is returned when user account is disabled
	ErrUserDisabled = errors.New("user account is disabled")
	// ErrInvalidToken is returned when JWT token is invalid
	ErrInvalidToken = errors.New("invalid token")
)

// JWTClaims represents JWT custom claims
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	DisplayName string `json:"display_name"`
	jwt.RegisteredClaims
}

// AuthManager handles JWT authentication operations
type AuthManager struct {
	db        store.Database
	jwtSecret []byte
}

// NewAuthManager creates a new auth manager
func NewAuthManager(db store.Database, jwtSecret string) *AuthManager {
	return &AuthManager{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// Login authenticates a user and returns a JWT token
func (am *AuthManager) Login(ctx context.Context, username, password string) (string, error) {
	// Get user from database
	user, err := am.db.GetAdminUserByUsername(ctx, username)
	if err != nil {
		if err == store.ErrNotFound {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return "", ErrUserDisabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := am.generateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// VerifyToken verifies a JWT token and returns the claims
func (am *AuthManager) VerifyToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ChangePassword changes a user's password
func (am *AuthManager) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	// Get user from database
	user, err := am.db.GetAdminUserByUsername(ctx, username)
	if err != nil {
		if err == store.ErrNotFound {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.Password = string(hashedPassword)
	if err := am.db.UpdateAdminUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// generateToken generates a JWT token for a user
func (am *AuthManager) generateToken(user *store.AdminUser) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		DisplayName: user.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

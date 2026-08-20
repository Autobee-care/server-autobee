// Package auth provides JWT token generation and validation.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// jwtClaims is the registered claim set stored inside each JWT.
type jwtClaims struct {
	jwt.RegisteredClaims
	TenantID string    `json:"tenantId"`
	Role     Role      `json:"role"`
	Type     TokenType `json:"type"`
}

// JWTService handles token generation and validation.
type JWTService struct {
	accessSecret      []byte
	refreshSecret     []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

// NewJWTService creates a JWTService with the given secrets and expiry durations.
func NewJWTService(accessSecret, refreshSecret string, accessExp, refreshExp time.Duration) *JWTService {
	return &JWTService{
		accessSecret:      []byte(accessSecret),
		refreshSecret:     []byte(refreshSecret),
		accessExpiration:  accessExp,
		refreshExpiration: refreshExp,
	}
}

// GenerateTokenPair creates both an access token and a refresh token for the given user.
func (s *JWTService) GenerateTokenPair(userID, tenantID string, role Role) (*TokenPair, error) {
	accessToken, err := s.generateToken(userID, tenantID, role, TokenTypeAccess, s.accessExpiration, s.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(userID, tenantID, role, TokenTypeRefresh, s.refreshExpiration, s.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken parses and validates an access token string.
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, TokenTypeAccess, s.accessSecret)
}

// ValidateRefreshToken parses and validates a refresh token string.
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, TokenTypeRefresh, s.refreshSecret)
}

func (s *JWTService) generateToken(
	userID, tenantID string,
	role Role,
	tokenType TokenType,
	expiration time.Duration,
	secret []byte,
) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			ID:        uuid.New().String(),
		},
		TenantID: tenantID,
		Role:     role,
		Type:     tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *JWTService) validateToken(tokenString string, expectedType TokenType, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.Type != expectedType {
		return nil, fmt.Errorf("wrong token type: expected %s, got %s", expectedType, claims.Type)
	}

	return &Claims{
		Subject:  claims.Subject,
		TenantID: claims.TenantID,
		Role:     claims.Role,
		Type:     claims.Type,
		JTI:      claims.ID,
	}, nil
}

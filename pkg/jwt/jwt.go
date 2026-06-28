package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType distinguishes access tokens from refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims is the JWT payload carried by both token types.
type Claims struct {
	UserID   uuid.UUID `json:"uid"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// Manager issues and verifies tokens with a shared secret (HS256).
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager builds a token Manager.
func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *Manager) generate(userID uuid.UUID, username, role string, t TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     t,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// GenerateAccess issues a short-lived access token.
func (m *Manager) GenerateAccess(userID uuid.UUID, username, role string) (string, error) {
	return m.generate(userID, username, role, AccessToken, m.accessTTL)
}

// GenerateRefresh issues a long-lived refresh token.
func (m *Manager) GenerateRefresh(userID uuid.UUID, username, role string) (string, error) {
	return m.generate(userID, username, role, RefreshToken, m.refreshTTL)
}

// Parse validates a token's signature and expiry and returns its claims.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

// AccessTTLSeconds exposes the access token lifetime in seconds (for clients).
func (m *Manager) AccessTTLSeconds() int {
	return int(m.accessTTL.Seconds())
}

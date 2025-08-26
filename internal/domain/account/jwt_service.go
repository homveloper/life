package account

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT token claims with UserID (not AccountID)
type JWTClaims struct {
	UserID string `json:"user_id"` // Game domain identifier
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey      []byte
	issuer         string
	expiryDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, issuer string, expiryDuration time.Duration) *JWTService {
	return &JWTService{
		secretKey:      []byte(secretKey),
		issuer:         issuer,
		expiryDuration: expiryDuration,
	}
}

// GenerateToken generates a new JWT token for an account (but contains UserID)
func (s *JWTService) GenerateToken(account *Account) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: account.UserID.String(), // Use UserID for game domain
		Email:  account.Profile.Email,
		Name:   account.Profile.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   account.UserID.String(), // Subject is UserID
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiryDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenMalformed
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

// RefreshToken creates a new token with extended expiry
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Create new token with fresh expiry
	now := time.Now()
	newClaims := JWTClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Name:   claims.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   claims.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiryDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(s.secretKey)
}
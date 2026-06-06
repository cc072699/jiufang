package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"jiufang/internal/pkg/config"
)

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenMalformed   = errors.New("token is malformed")
	ErrTokenSignatureInvalid = errors.New("token signature is invalid")
)

type Claims struct {
	UserID   int64   `json:"user_id"`
	Username string  `json:"username"`
	Role     string  `json:"role"`
	Groups   []int64 `json:"groups"`
	jwt.RegisteredClaims
}

type JWTManagerInterface interface {
	GenerateToken(userID int64, username, role string, groups []int64) (string, time.Time, error)
	ParseToken(tokenString string) (*Claims, error)
	ValidateToken(tokenString string) bool
}

type JWTManager struct {
	config *config.JWTConfig
}

func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{config: cfg}
}

func (m *JWTManager) GenerateToken(userID int64, username, role string, groups []int64) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.config.ExpireTime)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Groups:   groups,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "jiufang",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenSignatureInvalid
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, ErrTokenSignatureInvalid
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

func (m *JWTManager) ValidateToken(tokenString string) bool {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return false
	}
	return claims != nil
}
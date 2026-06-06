package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"jiufang/internal/pkg/jwt"
	"jiufang/internal/pkg/response"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UsernameKey         = "username"
	RoleKey             = "role"
	GroupsKey           = "groups"
)

type AuthMiddleware struct {
	jwtManager *jwt.JWTManager
}

func NewAuthMiddleware(jwtManager *jwt.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)
		if token == "" {
			response.Unauthorized(c, "missing token")
			c.Abort()
			return
		}

		claims, err := m.jwtManager.ParseToken(token)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				response.Error(c, 401, "token is expired")
				c.Abort()
				return
			}
			if err == jwt.ErrTokenMalformed || err == jwt.ErrTokenSignatureInvalid {
				response.Unauthorized(c, "invalid token")
				c.Abort()
				return
			}
			response.Unauthorized(c, "token validation failed")
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, claims.Role)
		c.Set(GroupsKey, claims.Groups)

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(RoleKey)
		if !exists {
			response.Unauthorized(c, "user role not found in context")
			c.Abort()
			return
		}

		role := userRole.(string)
		allowed := false
		for _, r := range roles {
			if role == r {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "permission denied")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a middleware that requires the user to have admin role.
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	return userID.(int64)
}

func GetUsername(c *gin.Context) string {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return ""
	}
	return username.(string)
}

func GetRole(c *gin.Context) string {
	role, exists := c.Get(RoleKey)
	if !exists {
		return ""
	}
	return role.(string)
}

func GetGroups(c *gin.Context) []int64 {
	groups, exists := c.Get(GroupsKey)
	if !exists {
		return []int64{}
	}
	return groups.([]int64)
}

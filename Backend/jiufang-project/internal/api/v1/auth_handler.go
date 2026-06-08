package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

type AuthHandler struct {
	authService *service.AuthAppService
}

func NewAuthHandler(authService *service.AuthAppService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	loginResp, err := h.authService.Login(c.Request.Context(), service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		if err == pkgerrors.ErrAccountNotFound {
			response.Error(c, http.StatusUnauthorized, "account not found")
			return
		}
		if err == pkgerrors.ErrUserDisabled {
			response.Error(c, http.StatusForbidden, "account has been disabled, please contact administrator")
			return
		}
		if err == pkgerrors.ErrInvalidCredentials {
			response.Error(c, http.StatusUnauthorized, "invalid username or password")
			return
		}
		response.InternalError(c, "login failed")
		return
	}

	// Set user_id in context for operation log middleware
	if uid, err := strconv.ParseInt(loginResp.User.ID, 10, 64); err == nil {
		c.Set("user_id", uid)
	}

	response.Success(c, gin.H{
		"token":      loginResp.Token,
		"expires_at": loginResp.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		"user": gin.H{
			"id":       loginResp.User.ID,
			"username": loginResp.User.Username,
			"role":     loginResp.User.Role,
			"groups":   loginResp.User.Groups,
			"email":    loginResp.User.Email,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	err := h.authService.Logout(c.Request.Context())
	if err != nil {
		response.InternalError(c, "logout failed")
		return
	}

	response.Success(c, nil)
}

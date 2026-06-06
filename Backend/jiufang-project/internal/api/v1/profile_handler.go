package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jiufang/internal/middleware"
	"jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/response"
	"jiufang/internal/pkg/upload"
	"jiufang/internal/service"
)

type ProfileHandler struct {
	profileService *service.ProfileAppService
}

func NewProfileHandler(profileService *service.ProfileAppService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	snowflakeID := middleware.GetUserID(c)
	if snowflakeID == 0 {
		response.Unauthorized(c, "unauthorized")
		return
	}

	u, err := h.profileService.GetProfile(c.Request.Context(), snowflakeID)
	if err != nil {
		if err == errors.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, "failed to get profile")
		return
	}

	groups := h.profileService.GetUserGroupIDs(c.Request.Context(), snowflakeID)

	response.Success(c, gin.H{
		"id":         strconv.FormatInt(u.SnowflakeID, 10),
		"username":   u.Username,
		"email":      u.Email,
		"role":       u.Role,
		"avatar":     u.Avatar,
		"groups":     groups,
		"status":     u.Status,
		"created_at": u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
	snowflakeID := middleware.GetUserID(c)
	if snowflakeID == 0 {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// Resolve snowflake ID to GORM primary key for avatar upload
	u, err := h.profileService.GetProfile(c.Request.Context(), snowflakeID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "avatar file is required")
		return
	}

	avatarURL, err := h.profileService.UploadAvatar(c.Request.Context(), u.ID, file)
	if err != nil {
		if err == upload.ErrInvalidFileType {
			response.BadRequest(c, "invalid file type, only JPG/PNG/GIF are allowed")
			return
		}
		if err == upload.ErrFileTooLarge {
			response.BadRequest(c, "file too large, maximum size is 2MB")
			return
		}
		response.InternalError(c, "failed to upload avatar")
		return
	}

	response.Success(c, gin.H{
		"avatar_url": avatarURL,
	})
}

func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	snowflakeID := middleware.GetUserID(c)
	if snowflakeID == 0 {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// Resolve snowflake ID to GORM primary key
	u, err := h.profileService.GetProfile(c.Request.Context(), snowflakeID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	err = h.profileService.ChangePassword(c.Request.Context(), u.ID, req.OldPassword, req.NewPassword, req.ConfirmPassword)
	if err != nil {
		if err == errors.ErrOldPasswordIncorrect {
			response.Error(c, http.StatusBadRequest, "current password is incorrect")
			return
		}
		if err == errors.ErrPasswordTooShort || err == errors.ErrPasswordTooLong {
			response.Error(c, http.StatusBadRequest, "password length must be between 6 and 20 characters")
			return
		}
		if err == errors.ErrPasswordNotMatch {
			response.Error(c, http.StatusBadRequest, "new password and confirm password do not match")
			return
		}
		if err == errors.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, "failed to change password")
		return
	}

	response.SuccessWithMessage(c, "密码修改成功，请重新登录", nil)
}
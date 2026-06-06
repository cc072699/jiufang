// Package v1 implements the HTTP handlers for user management.
package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/model/user"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	userService *service.UserAppService
	logger      *zap.Logger
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(userService *service.UserAppService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser handles POST /api/v1/users - create a new user.
func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse request
	var req user.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Create user
	newUser, err := h.userService.CreateUser(ctx, &req)
	if err != nil {
		if err == pkgerrors.ErrUserAlreadyExists {
			response.Error(c, http.StatusConflict, "username already exists")
			return
		}
		if err == pkgerrors.ErrEmailAlreadyExists {
			response.Error(c, http.StatusConflict, "email already exists")
			return
		}
		h.logger.Error("Failed to create user",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		response.InternalError(c, "failed to create user: "+err.Error())
		return
	}

	// Build response
	userResp := user.UserResponse{
		ID:        strconv.FormatInt(newUser.SnowflakeID, 10),
		Username:  newUser.Username,
		Email:     newUser.Email,
		Role:      user.Role(newUser.Role),
		Groups:    []string{}, // TODO: Get user groups
		Status:    newUser.Status,
		CreatedAt: newUser.CreatedAt,
	}

	response.Success(c, userResp)
}

// GetUser handles GET /api/v1/users/:id - get user by ID.
func (h *UserHandler) GetUser(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from path parameter
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id format")
		return
	}

	// Get user
	u, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		if err == pkgerrors.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		h.logger.Error("Failed to get user",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to get user: "+err.Error())
		return
	}

	// Build response
	userResp := user.UserResponse{
		ID:        strconv.FormatInt(u.SnowflakeID, 10),
		Username:  u.Username,
		Email:     u.Email,
		Role:      user.Role(u.Role),
		Groups:    []string{}, // TODO: Get user groups
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	response.Success(c, userResp)
}

// ListUsers handles GET /api/v1/users - list users with pagination and filters.
func (h *UserHandler) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	username := c.Query("username")
	role := c.Query("role")
	statusStr := c.Query("status")

	var status int
	if statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	} else {
		status = -1 // -1 means no filter
	}

	// Build request
	req := &user.ListUsersRequest{
		Page:     page,
		PageSize: pageSize,
		Username: username,
		Role:     user.Role(role),
		Status:   status,
	}

	// List users
	users, total, err := h.userService.ListUsers(ctx, req)
	if err != nil {
		h.logger.Error("Failed to list users",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Error(err),
		)
		response.InternalError(c, "failed to list users: "+err.Error())
		return
	}

	// Build response
	userList := make([]user.UserResponse, 0, len(users))
	for _, u := range users {
		userList = append(userList, user.UserResponse{
			ID:        strconv.FormatInt(u.SnowflakeID, 10),
			Username:  u.Username,
			Email:     u.Email,
			Role:      user.Role(u.Role),
			Groups:    []string{}, // TODO: Get user groups
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}

	response.PageWithField(c, "users", userList, total, page, pageSize)
}

// UpdateUser handles PUT /api/v1/users/:id - update user by ID.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from path parameter
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id format")
		return
	}

	// Parse request
	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Update user
	updatedUser, err := h.userService.UpdateUser(ctx, userID, &req)
	if err != nil {
		if err == pkgerrors.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		if err == pkgerrors.ErrUserAlreadyExists {
			response.Error(c, http.StatusConflict, "username already exists")
			return
		}
		if err == pkgerrors.ErrEmailAlreadyExists {
			response.Error(c, http.StatusConflict, "email already exists")
			return
		}
		h.logger.Error("Failed to update user",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to update user: "+err.Error())
		return
	}

	// Build response
	userResp := user.UserResponse{
		ID:        strconv.FormatInt(updatedUser.SnowflakeID, 10),
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		Role:      user.Role(updatedUser.Role),
		Groups:    []string{}, // TODO: Get user groups
		Status:    updatedUser.Status,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}

	response.Success(c, userResp)
}

// DeleteUser handles DELETE /api/v1/users/:id - delete user by ID.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from path parameter
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id format")
		return
	}

	// Delete user
	if err := h.userService.DeleteUser(ctx, userID); err != nil {
		if err == pkgerrors.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		h.logger.Error("Failed to delete user",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to delete user: "+err.Error())
		return
	}

	// Return success response
	response.Success(c, nil)
}

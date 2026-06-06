// Package service implements the application layer for user management.
package service

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"jiufang/internal/model/user"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// UserAppService manages user operations.
type UserAppService struct {
	userRepo    repository.UserRepositoryInterface
	groupRepo   repository.UserGroupRepositoryInterface
	idGenerator id.SnowflakeGeneratorInterface
	logger      *zap.Logger
}

// NewUserAppService creates a new UserAppService instance.
func NewUserAppService(
	userRepo repository.UserRepositoryInterface,
	groupRepo repository.UserGroupRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
	logger *zap.Logger,
) *UserAppService {
	return &UserAppService{
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateUser creates a new user.
func (s *UserAppService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	// Check if username already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("failed to check username existence",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if existingUser != nil {
		return nil, pkgerrors.ErrUserAlreadyExists
	}

	// Check if email already exists
	existingEmail, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error("failed to check email existence",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingEmail != nil {
		return nil, pkgerrors.ErrEmailAlreadyExists
	}

	// Generate snowflake ID
	snowflakeID := s.idGenerator.Generate()

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	newUser := &user.User{
		SnowflakeID: snowflakeID,
		Username:    req.Username,
		Password:    string(hashedPassword),
		Email:       req.Email,
		Role:        string(req.Role),
		Status:      1, // Default to active
	}

	// Save to database
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		s.logger.Error("failed to create user",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: Assign user to groups if specified in req.Groups

	s.logger.Info("user created successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.String("username", req.Username),
	)

	return newUser, nil
}

// GetUser retrieves a user by snowflake ID.
func (s *UserAppService) GetUser(ctx context.Context, snowflakeID int64) (*user.User, error) {
	u, err := s.userRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get user",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if u == nil {
		return nil, pkgerrors.ErrUserNotFound
	}

	return u, nil
}

// ListUsers retrieves a list of users with pagination and filters.
func (s *UserAppService) ListUsers(ctx context.Context, req *user.ListUsersRequest) ([]user.User, int64, error) {
	// Validate pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Convert status filter
	statusFilter := -1 // -1 means no filter
	if req.Status == 0 || req.Status == 1 {
		statusFilter = req.Status
	}

	users, total, err := s.userRepo.List(ctx, offset, req.PageSize, req.Username, string(req.Role), statusFilter)
	if err != nil {
		s.logger.Error("failed to list users",
			zap.Int("page", req.Page),
			zap.Int("page_size", req.PageSize),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates a user by snowflake ID.
func (s *UserAppService) UpdateUser(ctx context.Context, snowflakeID int64, req *user.UpdateUserRequest) (*user.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get user for update",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, pkgerrors.ErrUserNotFound
	}

	// Check username uniqueness if changing
	if req.Username != "" && req.Username != existingUser.Username {
		usernameExists, err := s.userRepo.GetByUsername(ctx, req.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to check username existence: %w", err)
		}
		if usernameExists != nil {
			return nil, pkgerrors.ErrUserAlreadyExists
		}
		existingUser.Username = req.Username
	}

	// Check email uniqueness if changing
	if req.Email != "" && req.Email != existingUser.Email {
		emailExists, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if emailExists != nil {
			return nil, pkgerrors.ErrEmailAlreadyExists
		}
		existingUser.Email = req.Email
	}

	// Update password if provided
	if req.Password != "" {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		existingUser.Password = string(hashedPwd)
	}

	// Update role if provided
	if req.Role != "" {
		existingUser.Role = string(req.Role)
	}

	// Update status if provided
	if req.Status == 0 || req.Status == 1 {
		existingUser.Status = req.Status
	}

	// TODO: Update user groups if specified in req.Groups

	// Save to database
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		s.logger.Error("failed to update user",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Info("user updated successfully",
		zap.Int64("snowflake_id", snowflakeID),
	)

	return existingUser, nil
}

// DeleteUser deletes a user by snowflake ID.
func (s *UserAppService) DeleteUser(ctx context.Context, snowflakeID int64) error {
	// Check if user exists
	existingUser, err := s.userRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get user for deletion",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return pkgerrors.ErrUserNotFound
	}

	// Delete user by ID (not snowflake ID)
	if err := s.userRepo.Delete(ctx, existingUser.ID); err != nil {
		s.logger.Error("failed to delete user",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Uint("id", existingUser.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Info("user deleted successfully",
		zap.Int64("snowflake_id", snowflakeID),
	)

	return nil
}

// GetUserGroups retrieves the group IDs (as strings) that a user belongs to.
func (s *UserAppService) GetUserGroups(ctx context.Context, snowflakeID int64) ([]string, error) {
	u, err := s.userRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil || u == nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	groups, err := s.groupRepo.GetGroupsByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	result := make([]string, 0, len(groups))
	for _, g := range groups {
		result = append(result, strconv.FormatInt(g.SnowflakeID, 10))
	}
	return result, nil
}

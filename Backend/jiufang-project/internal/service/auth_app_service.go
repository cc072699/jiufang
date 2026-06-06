package service

import (
	"context"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/jwt"
	"jiufang/internal/repository"
)

type AuthAppService struct {
	userRepo   repository.UserRepositoryInterface
	groupRepo  repository.UserGroupRepositoryInterface
	jwtManager jwt.JWTManagerInterface
}

func NewAuthAppService(userRepo repository.UserRepositoryInterface, groupRepo repository.UserGroupRepositoryInterface, jwtManager jwt.JWTManagerInterface) *AuthAppService {
	return &AuthAppService{
		userRepo:   userRepo,
		groupRepo:  groupRepo,
		jwtManager: jwtManager,
	}
}

type LoginRequest struct {
	Username string
	Password string
}

type LoginResponse struct {
	Token        string
	ExpiresAt    time.Time
	User         UserInfo
	IsFirstLogin bool
}

type UserInfo struct {
	ID       string // Snowflake ID as string
	Username string
	Role     string
	Groups   []string // Snowflake IDs as strings
}

func (s *AuthAppService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, pkgerrors.ErrAccountNotFound
	}

	if user.Status != 1 {
		return nil, pkgerrors.ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, pkgerrors.ErrInvalidCredentials
	}

	members, err := s.groupRepo.GetMembers(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]int64, 0)
	for _, member := range members {
		group, err := s.groupRepo.GetByID(ctx, member.GroupID)
		if err != nil {
			continue
		}
		if group != nil {
			groupIDs = append(groupIDs, group.SnowflakeID)
		}
	}

	// Convert snowflake IDs to strings for JSON response
	groupIDStrings := make([]string, len(groupIDs))
	for i, id := range groupIDs {
		groupIDStrings[i] = strconv.FormatInt(id, 10)
	}

	token, expiresAt, err := s.jwtManager.GenerateToken(user.SnowflakeID, user.Username, user.Role, groupIDs)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:        token,
		ExpiresAt:    expiresAt,
		IsFirstLogin: user.IsFirstLogin,
		User: UserInfo{
			ID:       strconv.FormatInt(user.SnowflakeID, 10),
			Username: user.Username,
			Role:     user.Role,
			Groups:   groupIDStrings,
		},
	}, nil
}

func (s *AuthAppService) Logout(ctx context.Context) error {
	return nil
}

func (s *AuthAppService) ValidateToken(token string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ParseToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (s *AuthAppService) MarkFirstLoginComplete(ctx context.Context, userID uint) error {
	return s.userRepo.UpdateFirstLoginStatus(ctx, userID, false)
}

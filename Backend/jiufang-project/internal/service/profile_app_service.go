package service

import (
	"context"
	"mime/multipart"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"jiufang/internal/model/user"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/upload"
	"jiufang/internal/repository"
)

type ProfileAppService struct {
	userRepo  repository.UserRepositoryInterface
	groupRepo repository.UserGroupRepositoryInterface
}

func NewProfileAppService(userRepo repository.UserRepositoryInterface, groupRepo repository.UserGroupRepositoryInterface) *ProfileAppService {
	return &ProfileAppService{userRepo: userRepo, groupRepo: groupRepo}
}

func (s *ProfileAppService) GetProfile(ctx context.Context, snowflakeID int64) (*user.User, error) {
	u, err := s.userRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, pkgerrors.ErrUserNotFound
	}
	return u, nil
}

// GetUserGroupIDs returns the user's group snowflake IDs as strings.
func (s *ProfileAppService) GetUserGroupIDs(ctx context.Context, snowflakeID int64) []string {
	u, err := s.userRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil || u == nil {
		return []string{}
	}
	groups, err := s.groupRepo.GetGroupsByUserID(ctx, u.ID)
	if err != nil || len(groups) == 0 {
		return []string{}
	}
	ids := make([]string, len(groups))
	for i, g := range groups {
		ids[i] = strconv.FormatInt(g.SnowflakeID, 10)
	}
	return ids
}

func (s *ProfileAppService) UploadAvatar(ctx context.Context, userID uint, file *multipart.FileHeader) (string, error) {
	result, err := upload.UploadAvatar(file, userID)
	if err != nil {
		return "", err
	}

	if err := s.userRepo.UpdateAvatar(ctx, userID, result.URL); err != nil {
		return "", err
	}

	return result.URL, nil
}

func (s *ProfileAppService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword, confirmPassword string) error {
	if len(newPassword) < 6 {
		return pkgerrors.ErrPasswordTooShort
	}
	if len(newPassword) > 20 {
		return pkgerrors.ErrPasswordTooLong
	}
	if newPassword != confirmPassword {
		return pkgerrors.ErrPasswordNotMatch
	}

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return pkgerrors.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPassword)); err != nil {
		return pkgerrors.ErrOldPasswordIncorrect
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"jiufang/internal/mocks"
	"jiufang/internal/model/user"
	pkgerrors "jiufang/internal/pkg/errors"
)

func TestProfileAppService_GetProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		SnowflakeID: 123456789,
		Username:    "testuser",
		Email:       "test@example.com",
		Avatar:      "",
		Role:        "admin",
		Status:      1,
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(expectedUser, nil)

	result, err := service.GetProfile(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
}

func TestProfileAppService_GetProfile_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	mockRepo.EXPECT().GetByID(gomock.Any(), uint(999)).Return(nil, nil)

	result, err := service.GetProfile(context.Background(), 999)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrUserNotFound, err)
	assert.Nil(t, result)
}

func TestProfileAppService_GetProfile_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	mockRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(nil, errors.New("database error"))

	result, err := service.GetProfile(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProfileAppService_ChangePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		Password: string(hashedPassword),
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(expectedUser, nil)
	mockRepo.EXPECT().UpdatePassword(gomock.Any(), uint(1), gomock.Any()).Return(nil)

	err := service.ChangePassword(context.Background(), 1, "oldpassword", "newpassword", "newpassword")

	assert.NoError(t, err)
}

func TestProfileAppService_ChangePassword_OldPasswordIncorrect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		Password: string(hashedPassword),
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(expectedUser, nil)

	err := service.ChangePassword(context.Background(), 1, "wrongpassword", "newpassword", "newpassword")

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrOldPasswordIncorrect, err)
}

func TestProfileAppService_ChangePassword_PasswordTooShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	err := service.ChangePassword(context.Background(), 1, "oldpassword", "12345", "12345")

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrPasswordTooShort, err)
}

func TestProfileAppService_ChangePassword_PasswordTooLong(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	longPassword := "thispasswordistoolongandexceedstwentycharacters"
	err := service.ChangePassword(context.Background(), 1, "oldpassword", longPassword, longPassword)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrPasswordTooLong, err)
}

func TestProfileAppService_ChangePassword_PasswordNotMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	err := service.ChangePassword(context.Background(), 1, "oldpassword", "newpassword", "differentpassword")

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrPasswordNotMatch, err)
}

func TestProfileAppService_ChangePassword_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	service := NewProfileAppService(mockRepo)

	mockRepo.EXPECT().GetByID(gomock.Any(), uint(999)).Return(nil, nil)

	err := service.ChangePassword(context.Background(), 999, "oldpassword", "newpassword", "newpassword")

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrUserNotFound, err)
}
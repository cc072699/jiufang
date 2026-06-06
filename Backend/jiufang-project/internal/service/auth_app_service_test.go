package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"jiufang/internal/mocks"
	"jiufang/internal/model/permission"
	"jiufang/internal/model/user"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/jwt"
)

func TestAuthAppService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		SnowflakeID:  123456789,
		Username:     "testuser",
		Password:     string(hashedPassword),
		Email:        "test@example.com",
		Role:         "admin",
		Status:       1,
		IsFirstLogin: false,
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)
	mockGroupRepo.EXPECT().GetMembers(gomock.Any(), uint(1)).Return([]user.UserGroupMember{}, nil)

	expectedExpiresAt := time.Now().Add(24 * time.Hour)
	mockJWTManager.EXPECT().GenerateToken(int64(123456789), "testuser", "admin", []int64{}).Return("test-token", expectedExpiresAt, nil)

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "testuser",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-token", result.Token)
	assert.Equal(t, "123456789", result.User.ID) // Snowflake ID as string
	assert.Equal(t, "testuser", result.User.Username)
	assert.Equal(t, "admin", result.User.Role)
	assert.False(t, result.IsFirstLogin)
}

func TestAuthAppService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "nonexistent").Return(nil, nil)

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrAccountNotFound, err)
	assert.Nil(t, result)
}

func TestAuthAppService_Login_UserDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		SnowflakeID: 123456789,
		Username:    "testuser",
		Password:    string(hashedPassword),
		Status:      0,
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "testuser",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrUserDisabled, err)
	assert.Nil(t, result)
}

func TestAuthAppService_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		SnowflakeID: 123456789,
		Username:    "testuser",
		Password:    string(hashedPassword),
		Status:      1,
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrInvalidCredentials, err)
	assert.Nil(t, result)
}

func TestAuthAppService_Login_WithGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		SnowflakeID:  123456789,
		Username:     "testuser",
		Password:     string(hashedPassword),
		Role:         "admin",
		Status:       1,
		IsFirstLogin: true,
	}

	expectedGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "管理员组",
	}

	members := []user.UserGroupMember{
		{UserID: 1, GroupID: 1},
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)
	mockGroupRepo.EXPECT().GetMembers(gomock.Any(), uint(1)).Return(members, nil)
	mockGroupRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(expectedGroup, nil)

	expectedExpiresAt := time.Now().Add(24 * time.Hour)
	mockJWTManager.EXPECT().GenerateToken(int64(123456789), "testuser", "admin", []int64{1000000000000000001}).Return("test-token", expectedExpiresAt, nil)

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "testuser",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsFirstLogin)
	assert.Equal(t, []string{"1000000000000000001"}, result.User.Groups) // Snowflake IDs as strings
}

func TestAuthAppService_Login_GetMembersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		SnowflakeID: 123456789,
		Username:    "testuser",
		Password:    string(hashedPassword),
		Status:      1,
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)
	mockGroupRepo.EXPECT().GetMembers(gomock.Any(), uint(1)).Return(nil, errors.New("database error"))

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "testuser",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthAppService_Logout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	err := service.Logout(context.Background())

	assert.NoError(t, err)
}

func TestAuthAppService_ValidateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	expectedClaims := &jwt.Claims{
		UserID:   123456789,
		Username: "testuser",
		Role:     "admin",
		Groups:   []int64{},
	}

	mockJWTManager.EXPECT().ParseToken("valid-token").Return(expectedClaims, nil)

	claims, err := service.ValidateToken("valid-token")

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, int64(123456789), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestAuthAppService_ValidateToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	mockJWTManager.EXPECT().ParseToken("invalid-token").Return(nil, jwt.ErrTokenInvalid)

	claims, err := service.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthAppService_ValidateToken_ExpiredToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	mockJWTManager.EXPECT().ParseToken("expired-token").Return(nil, jwt.ErrTokenExpired)

	claims, err := service.ValidateToken("expired-token")

	assert.Error(t, err)
	assert.Equal(t, jwt.ErrTokenExpired, err)
	assert.Nil(t, claims)
}

func TestAuthAppService_MarkFirstLoginComplete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockJWTManager := mocks.NewMockJWTManager(ctrl)

	service := NewAuthAppService(mockUserRepo, mockGroupRepo, mockJWTManager)

	mockUserRepo.EXPECT().UpdateFirstLoginStatus(gomock.Any(), uint(1), false).Return(nil)

	err := service.MarkFirstLoginComplete(context.Background(), 1)

	assert.NoError(t, err)
}

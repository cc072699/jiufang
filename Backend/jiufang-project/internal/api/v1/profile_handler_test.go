package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"jiufang/internal/mocks"
	"jiufang/internal/model/user"
	"jiufang/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestProfileHandler_GetProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	profileService := service.NewProfileAppService(mockRepo)
	handler := NewProfileHandler(profileService)

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

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(1))
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProfileHandler_GetProfile_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	profileService := service.NewProfileAppService(mockRepo)
	handler := NewProfileHandler(profileService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProfileHandler_ChangePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	profileService := service.NewProfileAppService(mockRepo)
	handler := NewProfileHandler(profileService)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	expectedUser := &user.User{
		Model: user.Model{
			ID: 1,
		},
		Password: string(hashedPassword),
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(expectedUser, nil)
	mockRepo.EXPECT().UpdatePassword(gomock.Any(), uint(1), gomock.Any()).Return(nil)

	body := ChangePasswordRequest{
		OldPassword:     "oldpassword",
		NewPassword:     "newpassword",
		ConfirmPassword: "newpassword",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(1))
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/profile/password", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProfileHandler_ChangePassword_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	profileService := service.NewProfileAppService(mockRepo)
	handler := NewProfileHandler(profileService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/profile/password", nil)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProfileHandler_ChangePassword_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	profileService := service.NewProfileAppService(mockRepo)
	handler := NewProfileHandler(profileService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(1))
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/profile/password", bytes.NewBuffer([]byte("{}")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

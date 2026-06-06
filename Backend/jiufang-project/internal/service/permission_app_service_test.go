package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"jiufang/internal/mocks"
	"jiufang/internal/model/permission"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/id"
)

func init() {
	id.Init(1)
}

func TestPermissionAppService_ConfigurePermissions_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockPermissionRepo.EXPECT().DeleteByGroupID(gomock.Any(), uint(1)).Return(nil)
	mockPermissionRepo.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).Return(nil)

	permissions := []PermissionRequest{
		{
			TableName:       "users",
			FilterCondition: "{\"department\": \"IT\"}",
		},
		{
			TableName:       "orders",
			FilterCondition: "{\"status\": \"active\"}",
		},
	}

	result, err := service.ConfigurePermissions(context.Background(), 1000000000000000001, permissions)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestPermissionAppService_ConfigurePermissions_GroupNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	permissions := []PermissionRequest{
		{
			TableName:       "users",
			FilterCondition: "{\"department\": \"IT\"}",
		},
	}

	result, err := service.ConfigurePermissions(context.Background(), 999999, permissions)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
	assert.Nil(t, result)
}

func TestPermissionAppService_ConfigurePermissions_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockPermissionRepo.EXPECT().DeleteByGroupID(gomock.Any(), uint(1)).Return(errors.New("database error"))

	permissions := []PermissionRequest{
		{
			TableName:       "users",
			FilterCondition: "{\"department\": \"IT\"}",
		},
	}

	result, err := service.ConfigurePermissions(context.Background(), 1000000000000000001, permissions)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPermissionAppService_ConfigurePermissions_CreateBatchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockPermissionRepo.EXPECT().DeleteByGroupID(gomock.Any(), uint(1)).Return(nil)
	mockPermissionRepo.EXPECT().CreateBatch(gomock.Any(), gomock.Any()).Return(errors.New("database error"))

	permissions := []PermissionRequest{
		{
			TableName:       "users",
			FilterCondition: "{\"department\": \"IT\"}",
		},
	}

	result, err := service.ConfigurePermissions(context.Background(), 1000000000000000001, permissions)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPermissionAppService_ConfigurePermissions_EmptyPermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockPermissionRepo.EXPECT().DeleteByGroupID(gomock.Any(), uint(1)).Return(nil)

	result, err := service.ConfigurePermissions(context.Background(), 1000000000000000001, []PermissionRequest{})

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestPermissionAppService_GetPermissionsByGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	permissions := []permission.Permission{
		{
			ID:              1,
			SnowflakeID:     1111111111111111111,
			GroupID:         1,
			TableName:       "users",
			AllowedFields:   "[\"id\", \"name\", \"email\"]",
			FilterCondition: "department = 'IT'",
		},
		{
			ID:              2,
			SnowflakeID:     2222222222222222222,
			GroupID:         1,
			TableName:       "orders",
			AllowedFields:   "[\"id\", \"order_no\", \"amount\"]",
			FilterCondition: "status = 'active'",
		},
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockPermissionRepo.EXPECT().GetByGroupID(gomock.Any(), uint(1)).Return(permissions, nil)

	result, err := service.GetPermissionsByGroup(context.Background(), 1000000000000000001)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "users", result[0].TableName)
	assert.Equal(t, "department = 'IT'", result[0].FilterCondition)
}

func TestPermissionAppService_GetPermissionsByGroup_GroupNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	result, err := service.GetPermissionsByGroup(context.Background(), 999999)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
	assert.Nil(t, result)
}

func TestPermissionAppService_GetPermissionsByGroup_GetPermissionsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockPermissionRepo.EXPECT().GetByGroupID(gomock.Any(), uint(1)).Return(nil, errors.New("database error"))

	result, err := service.GetPermissionsByGroup(context.Background(), 1000000000000000001)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestPermissionAppService_GetPermissionsByGroup_EmptyPermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewPermissionAppService(mockGroupRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockPermissionRepo.EXPECT().GetByGroupID(gomock.Any(), uint(1)).Return([]permission.Permission{}, nil)

	result, err := service.GetPermissionsByGroup(context.Background(), 1000000000000000001)

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

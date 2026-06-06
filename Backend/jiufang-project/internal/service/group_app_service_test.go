package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"jiufang/internal/mocks"
	"jiufang/internal/model/permission"
	"jiufang/internal/model/user"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/id"
)

func init() {
	id.Init(1)
}

func TestGroupAppService_CreateGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetByName(gomock.Any(), "测试组").Return(nil, nil)

	mockGroupRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, group *permission.UserGroup) error {
		group.ID = 1
		return nil
	})

	result, err := service.CreateGroup(context.Background(), CreateGroupRequest{
		Name:        "测试组",
		Description: "这是一个测试组",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "测试组", result.Name)
	assert.Equal(t, "这是一个测试组", result.Description)
}

func TestGroupAppService_CreateGroup_WithMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetByName(gomock.Any(), "测试组").Return(nil, nil)

	mockGroupRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, group *permission.UserGroup) error {
		group.ID = 1
		return nil
	})

	expectedUser := &user.User{
		Model: user.Model{
			ID: 10,
		},
		SnowflakeID: 123456789,
		Username:    "testuser",
	}

	mockUserRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(123456789)).Return(expectedUser, nil)
	mockGroupRepo.EXPECT().AddMembers(gomock.Any(), uint(1), []uint{10}).Return(nil)

	result, err := service.CreateGroup(context.Background(), CreateGroupRequest{
		Name:        "测试组",
		Description: "这是一个测试组",
		MemberIDs:   []int64{123456789},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGroupAppService_CreateGroup_NameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:   1,
		Name: "已存在组",
	}

	mockGroupRepo.EXPECT().GetByName(gomock.Any(), "已存在组").Return(existingGroup, nil)

	result, err := service.CreateGroup(context.Background(), CreateGroupRequest{
		Name:        "已存在组",
		Description: "描述",
	})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNameExists, err)
	assert.Nil(t, result)
}

func TestGroupAppService_GetGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	expectedGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
		Description: "描述",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(expectedGroup, nil)
	mockGroupRepo.EXPECT().GetMemberCount(gomock.Any(), uint(1)).Return(int64(5), nil)
	mockGroupRepo.EXPECT().GetMembers(gomock.Any(), uint(1)).Return([]user.UserGroupMember{}, nil)

	result, err := service.GetGroup(context.Background(), 1000000000000000001)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "测试组", result.Name)
	assert.Equal(t, int64(5), result.MemberCount)
}

func TestGroupAppService_GetGroup_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	result, err := service.GetGroup(context.Background(), 999999)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
	assert.Nil(t, result)
}

func TestGroupAppService_ListGroups_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	groups := []permission.UserGroup{
		{ID: 1, SnowflakeID: 1000000000000000001, Name: "组1"},
		{ID: 2, SnowflakeID: 1000000000000000002, Name: "组2"},
	}

	mockGroupRepo.EXPECT().List(gomock.Any(), 0, 10, "").Return(groups, int64(2), nil)
	mockGroupRepo.EXPECT().GetMemberCount(gomock.Any(), uint(1)).Return(int64(3), nil)
	mockGroupRepo.EXPECT().GetMemberCount(gomock.Any(), uint(2)).Return(int64(5), nil)

	result, total, err := service.ListGroups(context.Background(), 1, 10, "")

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
}

func TestGroupAppService_UpdateGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "旧名称",
		Description: "旧描述",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().GetByName(gomock.Any(), "新名称").Return(nil, nil)
	mockGroupRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	result, err := service.UpdateGroup(context.Background(), 1000000000000000001, UpdateGroupRequest{
		Name:        "新名称",
		Description: "新描述",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGroupAppService_UpdateGroup_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	result, err := service.UpdateGroup(context.Background(), 999999, UpdateGroupRequest{
		Name: "新名称",
	})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
	assert.Nil(t, result)
}

func TestGroupAppService_UpdateGroup_NameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "旧名称",
	}

	otherGroup := &permission.UserGroup{
		ID:          2,
		SnowflakeID: 1000000000000000002,
		Name:        "已存在名称",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().GetByName(gomock.Any(), "已存在名称").Return(otherGroup, nil)

	result, err := service.UpdateGroup(context.Background(), 1000000000000000001, UpdateGroupRequest{
		Name: "已存在名称",
	})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNameExists, err)
	assert.Nil(t, result)
}

func TestGroupAppService_DeleteGroup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "可删除组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().IsPresetGroup(gomock.Any(), int64(1000000000000000001)).Return(false, nil)
	mockGroupRepo.EXPECT().Delete(gomock.Any(), uint(1)).Return(nil)

	err := service.DeleteGroup(context.Background(), 1000000000000000001)

	assert.NoError(t, err)
}

func TestGroupAppService_DeleteGroup_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	err := service.DeleteGroup(context.Background(), 999999)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
}

func TestGroupAppService_DeleteGroup_PresetGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "预置组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().IsPresetGroup(gomock.Any(), int64(1000000000000000001)).Return(true, nil)

	err := service.DeleteGroup(context.Background(), 1000000000000000001)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrPresetGroupCannotDelete, err)
}

func TestGroupAppService_UpdateGroup_WithMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	expectedUser := &user.User{
		Model: user.Model{
			ID: 10,
		},
		SnowflakeID: 123456789,
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().GetMembers(gomock.Any(), uint(1)).Return([]user.UserGroupMember{}, nil)
	mockUserRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(123456789)).Return(expectedUser, nil)
	mockGroupRepo.EXPECT().AddMembers(gomock.Any(), uint(1), []uint{10}).Return(nil)
	mockGroupRepo.EXPECT().Update(gomock.Any(), existingGroup).Return(nil)

	result, err := service.UpdateGroup(context.Background(), 1000000000000000001, UpdateGroupRequest{
		MemberIDs: []int64{123456789},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGroupAppService_GetGroup_WithMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	members := []user.UserGroupMember{
		{UserID: 1, GroupID: 1},
		{UserID: 2, GroupID: 1},
	}

	user1 := &user.User{Model: user.Model{ID: 1}, SnowflakeID: 111, Username: "user1"}
	user2 := &user.User{Model: user.Model{ID: 2}, SnowflakeID: 222, Username: "user2"}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().GetMemberCount(gomock.Any(), uint(1)).Return(int64(2), nil)
	mockGroupRepo.EXPECT().GetMembers(gomock.Any(), uint(1)).Return(members, nil)
	mockUserRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(user1, nil)
	mockUserRepo.EXPECT().GetByID(gomock.Any(), uint(2)).Return(user2, nil)

	result, err := service.GetGroup(context.Background(), 1000000000000000001)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Members, 2)
}

func TestGroupAppService_GetGroup_GetMembersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().GetMemberCount(gomock.Any(), uint(1)).Return(int64(0), nil)
	mockGroupRepo.EXPECT().GetMembers(gomock.Any(), uint(1)).Return(nil, errors.New("database error"))

	result, err := service.GetGroup(context.Background(), 1000000000000000001)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGroupAppService_GetGroupMembers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
		Name:        "测试组",
	}

	members := []user.UserGroupMember{
		{UserID: 1, GroupID: 1},
		{UserID: 2, GroupID: 1},
	}

	user1 := &user.User{Model: user.Model{ID: 1}, SnowflakeID: 111, Username: "user1", Email: "user1@example.com"}
	user2 := &user.User{Model: user.Model{ID: 2}, SnowflakeID: 222, Username: "user2", Email: "user2@example.com"}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().GetMembersWithPagination(gomock.Any(), uint(1), 0, 20).Return(members, int64(2), nil)
	mockUserRepo.EXPECT().GetByID(gomock.Any(), uint(1)).Return(user1, nil)
	mockUserRepo.EXPECT().GetByID(gomock.Any(), uint(2)).Return(user2, nil)

	result, total, err := service.GetGroupMembers(context.Background(), 1000000000000000001, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.Equal(t, "user1", result[0].Username)
	assert.Equal(t, "user2", result[1].Username)
}

func TestGroupAppService_GetGroupMembers_GroupNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	result, total, err := service.GetGroupMembers(context.Background(), 999999, 1, 20)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
}

func TestGroupAppService_GetGroupMembers_InvalidPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockGroupRepo.EXPECT().GetMembersWithPagination(gomock.Any(), uint(1), 0, 20).Return([]user.UserGroupMember{}, int64(0), nil)

	result, total, err := service.GetGroupMembers(context.Background(), 1000000000000000001, 0, 150)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, result, 0)
}

func TestGroupAppService_AddGroupMembers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
	}

	user1 := &user.User{Model: user.Model{ID: 1}, SnowflakeID: 111, Username: "user1", Email: "user1@example.com"}
	user2 := &user.User{Model: user.Model{ID: 2}, SnowflakeID: 222, Username: "user2", Email: "user2@example.com"}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockUserRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(111)).Return(user1, nil)
	mockUserRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(222)).Return(user2, nil)
	mockGroupRepo.EXPECT().AddMembers(gomock.Any(), uint(1), []uint{1, 2}).Return(nil)

	result, err := service.AddGroupMembers(context.Background(), 1000000000000000001, []int64{111, 222})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "user1", result[0].Username)
	assert.Equal(t, "user2", result[1].Username)
}

func TestGroupAppService_AddGroupMembers_GroupNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	result, err := service.AddGroupMembers(context.Background(), 999999, []int64{111})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
	assert.Nil(t, result)
}

func TestGroupAppService_AddGroupMembers_InvalidUserIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockUserRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	result, err := service.AddGroupMembers(context.Background(), 1000000000000000001, []int64{999999})

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrInvalidRequest, err)
	assert.Nil(t, result)
}

func TestGroupAppService_RemoveGroupMember_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
	}

	user := &user.User{Model: user.Model{ID: 1}, SnowflakeID: 111}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockUserRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(111)).Return(user, nil)
	mockGroupRepo.EXPECT().RemoveMembers(gomock.Any(), uint(1), []uint{1}).Return(nil)

	err := service.RemoveGroupMember(context.Background(), 1000000000000000001, 111)

	assert.NoError(t, err)
}

func TestGroupAppService_RemoveGroupMember_GroupNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	err := service.RemoveGroupMember(context.Background(), 999999, 111)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrGroupNotFound, err)
}

func TestGroupAppService_RemoveGroupMember_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGroupRepo := mocks.NewMockUserGroupRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockPermissionRepo := mocks.NewMockPermissionRepository(ctrl)

	service := NewGroupAppService(mockGroupRepo, mockUserRepo, mockPermissionRepo)

	existingGroup := &permission.UserGroup{
		ID:          1,
		SnowflakeID: 1000000000000000001,
	}

	mockGroupRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(1000000000000000001)).Return(existingGroup, nil)
	mockUserRepo.EXPECT().GetBySnowflakeID(gomock.Any(), int64(999999)).Return(nil, nil)

	err := service.RemoveGroupMember(context.Background(), 1000000000000000001, 999999)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.ErrUserNotFound, err)
}

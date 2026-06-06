package mocks

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"

	"jiufang/internal/model/permission"
	"jiufang/internal/model/user"
)

type MockUserGroupRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserGroupRepositoryMockRecorder
}

type MockUserGroupRepositoryMockRecorder struct {
	mock *MockUserGroupRepository
}

func NewMockUserGroupRepository(ctrl *gomock.Controller) *MockUserGroupRepository {
	mock := &MockUserGroupRepository{ctrl: ctrl}
	mock.recorder = &MockUserGroupRepositoryMockRecorder{mock}
	return mock
}

func (m *MockUserGroupRepository) EXPECT() *MockUserGroupRepositoryMockRecorder {
	return m.recorder
}

func (m *MockUserGroupRepository) Create(ctx context.Context, group *permission.UserGroup) error {
	ret := m.ctrl.Call(m, "Create", ctx, group)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserGroupRepositoryMockRecorder) Create(ctx, group interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUserGroupRepository)(nil).Create), ctx, group)
}

func (m *MockUserGroupRepository) GetByID(ctx context.Context, id uint) (*permission.UserGroup, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*permission.UserGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserGroupRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUserGroupRepository)(nil).GetByID), ctx, id)
}

func (m *MockUserGroupRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*permission.UserGroup, error) {
	ret := m.ctrl.Call(m, "GetBySnowflakeID", ctx, snowflakeID)
	ret0, _ := ret[0].(*permission.UserGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserGroupRepositoryMockRecorder) GetBySnowflakeID(ctx, snowflakeID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySnowflakeID", reflect.TypeOf((*MockUserGroupRepository)(nil).GetBySnowflakeID), ctx, snowflakeID)
}

func (m *MockUserGroupRepository) GetByName(ctx context.Context, name string) (*permission.UserGroup, error) {
	ret := m.ctrl.Call(m, "GetByName", ctx, name)
	ret0, _ := ret[0].(*permission.UserGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserGroupRepositoryMockRecorder) GetByName(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByName", reflect.TypeOf((*MockUserGroupRepository)(nil).GetByName), ctx, name)
}

func (m *MockUserGroupRepository) List(ctx context.Context, offset, limit int, name string) ([]permission.UserGroup, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, offset, limit, name)
	ret0, _ := ret[0].([]permission.UserGroup)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockUserGroupRepositoryMockRecorder) List(ctx, offset, limit, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUserGroupRepository)(nil).List), ctx, offset, limit, name)
}

func (m *MockUserGroupRepository) Update(ctx context.Context, group *permission.UserGroup) error {
	ret := m.ctrl.Call(m, "Update", ctx, group)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserGroupRepositoryMockRecorder) Update(ctx, group interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUserGroupRepository)(nil).Update), ctx, group)
}

func (m *MockUserGroupRepository) Delete(ctx context.Context, id uint) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserGroupRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserGroupRepository)(nil).Delete), ctx, id)
}

func (m *MockUserGroupRepository) GetMemberCount(ctx context.Context, groupID uint) (int64, error) {
	ret := m.ctrl.Call(m, "GetMemberCount", ctx, groupID)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserGroupRepositoryMockRecorder) GetMemberCount(ctx, groupID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMemberCount", reflect.TypeOf((*MockUserGroupRepository)(nil).GetMemberCount), ctx, groupID)
}

func (m *MockUserGroupRepository) AddMembers(ctx context.Context, groupID uint, userIDs []uint) error {
	ret := m.ctrl.Call(m, "AddMembers", ctx, groupID, userIDs)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserGroupRepositoryMockRecorder) AddMembers(ctx, groupID, userIDs interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMembers", reflect.TypeOf((*MockUserGroupRepository)(nil).AddMembers), ctx, groupID, userIDs)
}

func (m *MockUserGroupRepository) RemoveMembers(ctx context.Context, groupID uint, userIDs []uint) error {
	ret := m.ctrl.Call(m, "RemoveMembers", ctx, groupID, userIDs)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserGroupRepositoryMockRecorder) RemoveMembers(ctx, groupID, userIDs interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveMembers", reflect.TypeOf((*MockUserGroupRepository)(nil).RemoveMembers), ctx, groupID, userIDs)
}

func (m *MockUserGroupRepository) GetMembers(ctx context.Context, groupID uint) ([]user.UserGroupMember, error) {
	ret := m.ctrl.Call(m, "GetMembers", ctx, groupID)
	ret0, _ := ret[0].([]user.UserGroupMember)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserGroupRepositoryMockRecorder) GetMembers(ctx, groupID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMembers", reflect.TypeOf((*MockUserGroupRepository)(nil).GetMembers), ctx, groupID)
}

func (m *MockUserGroupRepository) GetMembersWithPagination(ctx context.Context, groupID uint, offset, limit int) ([]user.UserGroupMember, int64, error) {
	ret := m.ctrl.Call(m, "GetMembersWithPagination", ctx, groupID, offset, limit)
	ret0, _ := ret[0].([]user.UserGroupMember)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockUserGroupRepositoryMockRecorder) GetMembersWithPagination(ctx, groupID, offset, limit interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMembersWithPagination", reflect.TypeOf((*MockUserGroupRepository)(nil).GetMembersWithPagination), ctx, groupID, offset, limit)
}

func (m *MockUserGroupRepository) GetGroupsByUserID(ctx context.Context, userID uint) ([]permission.UserGroup, error) {
	ret := m.ctrl.Call(m, "GetGroupsByUserID", ctx, userID)
	ret0, _ := ret[0].([]permission.UserGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserGroupRepositoryMockRecorder) GetGroupsByUserID(ctx, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupsByUserID", reflect.TypeOf((*MockUserGroupRepository)(nil).GetGroupsByUserID), ctx, userID)
}

func (m *MockUserGroupRepository) IsPresetGroup(ctx context.Context, snowflakeID int64) (bool, error) {
	ret := m.ctrl.Call(m, "IsPresetGroup", ctx, snowflakeID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserGroupRepositoryMockRecorder) IsPresetGroup(ctx, snowflakeID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPresetGroup", reflect.TypeOf((*MockUserGroupRepository)(nil).IsPresetGroup), ctx, snowflakeID)
}

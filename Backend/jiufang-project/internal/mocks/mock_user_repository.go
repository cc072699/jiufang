package mocks

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"

	"jiufang/internal/model/user"
)

type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	ret := m.ctrl.Call(m, "Create", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) Create(ctx, u interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUserRepository)(nil).Create), ctx, u)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*user.User, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUserRepository)(nil).GetByID), ctx, id)
}

func (m *MockUserRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*user.User, error) {
	ret := m.ctrl.Call(m, "GetBySnowflakeID", ctx, snowflakeID)
	ret0, _ := ret[0].(*user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetBySnowflakeID(ctx, snowflakeID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySnowflakeID", reflect.TypeOf((*MockUserRepository)(nil).GetBySnowflakeID), ctx, snowflakeID)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	ret := m.ctrl.Call(m, "GetByUsername", ctx, username)
	ret0, _ := ret[0].(*user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetByUsername(ctx, username interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUsername", reflect.TypeOf((*MockUserRepository)(nil).GetByUsername), ctx, username)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	ret := m.ctrl.Call(m, "GetByEmail", ctx, email)
	ret0, _ := ret[0].(*user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetByEmail(ctx, email interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByEmail", reflect.TypeOf((*MockUserRepository)(nil).GetByEmail), ctx, email)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int, username, role string, status int) ([]user.User, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, offset, limit, username, role, status)
	ret0, _ := ret[0].([]user.User)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockUserRepositoryMockRecorder) List(ctx, offset, limit, username, role, status interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUserRepository)(nil).List), ctx, offset, limit, username, role, status)
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	ret := m.ctrl.Call(m, "Update", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) Update(ctx, u interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUserRepository)(nil).Update), ctx, u)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserRepository)(nil).Delete), ctx, id)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, id uint, status int) error {
	ret := m.ctrl.Call(m, "UpdateStatus", ctx, id, status)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) UpdateStatus(ctx, id, status interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockUserRepository)(nil).UpdateStatus), ctx, id, status)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	ret := m.ctrl.Call(m, "UpdatePassword", ctx, id, password)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) UpdatePassword(ctx, id, password interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockUserRepository)(nil).UpdatePassword), ctx, id, password)
}

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, id uint, avatar string) error {
	ret := m.ctrl.Call(m, "UpdateAvatar", ctx, id, avatar)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) UpdateAvatar(ctx, id, avatar interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAvatar", reflect.TypeOf((*MockUserRepository)(nil).UpdateAvatar), ctx, id, avatar)
}

func (m *MockUserRepository) UpdateFirstLoginStatus(ctx context.Context, id uint, isFirstLogin bool) error {
	ret := m.ctrl.Call(m, "UpdateFirstLoginStatus", ctx, id, isFirstLogin)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) UpdateFirstLoginStatus(ctx, id, isFirstLogin interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFirstLoginStatus", reflect.TypeOf((*MockUserRepository)(nil).UpdateFirstLoginStatus), ctx, id, isFirstLogin)
}

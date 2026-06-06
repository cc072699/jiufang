package mocks

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"

	"jiufang/internal/model/permission"
)

type MockPermissionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockPermissionRepositoryMockRecorder
}

type MockPermissionRepositoryMockRecorder struct {
	mock *MockPermissionRepository
}

func NewMockPermissionRepository(ctrl *gomock.Controller) *MockPermissionRepository {
	mock := &MockPermissionRepository{ctrl: ctrl}
	mock.recorder = &MockPermissionRepositoryMockRecorder{mock}
	return mock
}

func (m *MockPermissionRepository) EXPECT() *MockPermissionRepositoryMockRecorder {
	return m.recorder
}

func (m *MockPermissionRepository) Create(ctx context.Context, p *permission.Permission) error {
	ret := m.ctrl.Call(m, "Create", ctx, p)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) Create(ctx, p interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockPermissionRepository)(nil).Create), ctx, p)
}

func (m *MockPermissionRepository) CreateBatch(ctx context.Context, permissions []permission.Permission) error {
	ret := m.ctrl.Call(m, "CreateBatch", ctx, permissions)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) CreateBatch(ctx, permissions interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBatch", reflect.TypeOf((*MockPermissionRepository)(nil).CreateBatch), ctx, permissions)
}

func (m *MockPermissionRepository) GetByID(ctx context.Context, id uint) (*permission.Permission, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*permission.Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPermissionRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockPermissionRepository)(nil).GetByID), ctx, id)
}

func (m *MockPermissionRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*permission.Permission, error) {
	ret := m.ctrl.Call(m, "GetBySnowflakeID", ctx, snowflakeID)
	ret0, _ := ret[0].(*permission.Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPermissionRepositoryMockRecorder) GetBySnowflakeID(ctx, snowflakeID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySnowflakeID", reflect.TypeOf((*MockPermissionRepository)(nil).GetBySnowflakeID), ctx, snowflakeID)
}

func (m *MockPermissionRepository) GetByGroupID(ctx context.Context, groupID uint) ([]permission.Permission, error) {
	ret := m.ctrl.Call(m, "GetByGroupID", ctx, groupID)
	ret0, _ := ret[0].([]permission.Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPermissionRepositoryMockRecorder) GetByGroupID(ctx, groupID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByGroupID", reflect.TypeOf((*MockPermissionRepository)(nil).GetByGroupID), ctx, groupID)
}

func (m *MockPermissionRepository) Update(ctx context.Context, p *permission.Permission) error {
	ret := m.ctrl.Call(m, "Update", ctx, p)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) Update(ctx, p interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockPermissionRepository)(nil).Update), ctx, p)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id uint) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockPermissionRepository)(nil).Delete), ctx, id)
}

func (m *MockPermissionRepository) DeleteByGroupID(ctx context.Context, groupID uint) error {
	ret := m.ctrl.Call(m, "DeleteByGroupID", ctx, groupID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) DeleteByGroupID(ctx, groupID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByGroupID", reflect.TypeOf((*MockPermissionRepository)(nil).DeleteByGroupID), ctx, groupID)
}

func (m *MockPermissionRepository) List(ctx context.Context, offset, limit int, groupID uint, resourceType string) ([]permission.Permission, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, offset, limit, groupID, resourceType)
	ret0, _ := ret[0].([]permission.Permission)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockPermissionRepositoryMockRecorder) List(ctx, offset, limit, groupID, resourceType interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPermissionRepository)(nil).List), ctx, offset, limit, groupID, resourceType)
}

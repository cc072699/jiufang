package mocks

import (
	"reflect"
	"time"

	"github.com/golang/mock/gomock"

	"jiufang/internal/pkg/jwt"
)

type MockJWTManager struct {
	ctrl     *gomock.Controller
	recorder *MockJWTManagerMockRecorder
}

type MockJWTManagerMockRecorder struct {
	mock *MockJWTManager
}

func NewMockJWTManager(ctrl *gomock.Controller) *MockJWTManager {
	mock := &MockJWTManager{ctrl: ctrl}
	mock.recorder = &MockJWTManagerMockRecorder{mock}
	return mock
}

func (m *MockJWTManager) EXPECT() *MockJWTManagerMockRecorder {
	return m.recorder
}

func (m *MockJWTManager) GenerateToken(userID int64, username, role string, groups []int64) (string, time.Time, error) {
	ret := m.ctrl.Call(m, "GenerateToken", userID, username, role, groups)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockJWTManagerMockRecorder) GenerateToken(userID, username, role, groups interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateToken", reflect.TypeOf((*MockJWTManager)(nil).GenerateToken), userID, username, role, groups)
}

func (m *MockJWTManager) ParseToken(tokenString string) (*jwt.Claims, error) {
	ret := m.ctrl.Call(m, "ParseToken", tokenString)
	ret0, _ := ret[0].(*jwt.Claims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockJWTManagerMockRecorder) ParseToken(tokenString interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseToken", reflect.TypeOf((*MockJWTManager)(nil).ParseToken), tokenString)
}

func (m *MockJWTManager) ValidateToken(tokenString string) bool {
	ret := m.ctrl.Call(m, "ValidateToken", tokenString)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (mr *MockJWTManagerMockRecorder) ValidateToken(tokenString interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken", reflect.TypeOf((*MockJWTManager)(nil).ValidateToken), tokenString)
}
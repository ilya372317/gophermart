// Code generated by MockGen. DO NOT EDIT.
// Source: internal/handler/register.go

// Package handler_mock is a generated GoMock package.
package handler_mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	entity "github.com/ilya372317/gophermart/internal/entity"
)

// MockRegisterStorage is a mock of RegisterStorage interface.
type MockRegisterStorage struct {
	ctrl     *gomock.Controller
	recorder *MockRegisterStorageMockRecorder
}

// MockRegisterStorageMockRecorder is the mock recorder for MockRegisterStorage.
type MockRegisterStorageMockRecorder struct {
	mock *MockRegisterStorage
}

// NewMockRegisterStorage creates a new mock instance.
func NewMockRegisterStorage(ctrl *gomock.Controller) *MockRegisterStorage {
	mock := &MockRegisterStorage{ctrl: ctrl}
	mock.recorder = &MockRegisterStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegisterStorage) EXPECT() *MockRegisterStorageMockRecorder {
	return m.recorder
}

// GetUserByLogin mocks base method.
func (m *MockRegisterStorage) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByLogin", ctx, login)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByLogin indicates an expected call of GetUserByLogin.
func (mr *MockRegisterStorageMockRecorder) GetUserByLogin(ctx, login interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByLogin", reflect.TypeOf((*MockRegisterStorage)(nil).GetUserByLogin), ctx, login)
}

// HasUser mocks base method.
func (m *MockRegisterStorage) HasUser(ctx context.Context, login string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasUser", ctx, login)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasUser indicates an expected call of HasUser.
func (mr *MockRegisterStorageMockRecorder) HasUser(ctx, login interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasUser", reflect.TypeOf((*MockRegisterStorage)(nil).HasUser), ctx, login)
}

// SaveUser mocks base method.
func (m *MockRegisterStorage) SaveUser(ctx context.Context, user entity.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveUser", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveUser indicates an expected call of SaveUser.
func (mr *MockRegisterStorageMockRecorder) SaveUser(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveUser", reflect.TypeOf((*MockRegisterStorage)(nil).SaveUser), ctx, user)
}

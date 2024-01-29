// Code generated by MockGen. DO NOT EDIT.
// Source: internal/handler/orderreg.go

// Package handler_mock is a generated GoMock package.
package handler_mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	entity "github.com/ilya372317/gophermart/internal/entity"
)

// MockRegisterOrderStorage is a mock of RegisterOrderStorage interface.
type MockRegisterOrderStorage struct {
	ctrl     *gomock.Controller
	recorder *MockRegisterOrderStorageMockRecorder
}

// MockRegisterOrderStorageMockRecorder is the mock recorder for MockRegisterOrderStorage.
type MockRegisterOrderStorageMockRecorder struct {
	mock *MockRegisterOrderStorage
}

// NewMockRegisterOrderStorage creates a new mock instance.
func NewMockRegisterOrderStorage(ctrl *gomock.Controller) *MockRegisterOrderStorage {
	mock := &MockRegisterOrderStorage{ctrl: ctrl}
	mock.recorder = &MockRegisterOrderStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRegisterOrderStorage) EXPECT() *MockRegisterOrderStorageMockRecorder {
	return m.recorder
}

// GetOrderByNumber mocks base method.
func (m *MockRegisterOrderStorage) GetOrderByNumber(arg0 context.Context, arg1 int) (*entity.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrderByNumber", arg0, arg1)
	ret0, _ := ret[0].(*entity.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrderByNumber indicates an expected call of GetOrderByNumber.
func (mr *MockRegisterOrderStorageMockRecorder) GetOrderByNumber(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrderByNumber", reflect.TypeOf((*MockRegisterOrderStorage)(nil).GetOrderByNumber), arg0, arg1)
}

// HasOrderByNumber mocks base method.
func (m *MockRegisterOrderStorage) HasOrderByNumber(arg0 context.Context, arg1 int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasOrderByNumber", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasOrderByNumber indicates an expected call of HasOrderByNumber.
func (mr *MockRegisterOrderStorageMockRecorder) HasOrderByNumber(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasOrderByNumber", reflect.TypeOf((*MockRegisterOrderStorage)(nil).HasOrderByNumber), arg0, arg1)
}

// HasOrderByNumberAndUserID mocks base method.
func (m *MockRegisterOrderStorage) HasOrderByNumberAndUserID(arg0 context.Context, arg1 int, arg2 uint) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasOrderByNumberAndUserID", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasOrderByNumberAndUserID indicates an expected call of HasOrderByNumberAndUserID.
func (mr *MockRegisterOrderStorageMockRecorder) HasOrderByNumberAndUserID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasOrderByNumberAndUserID", reflect.TypeOf((*MockRegisterOrderStorage)(nil).HasOrderByNumberAndUserID), arg0, arg1, arg2)
}

// SaveOrder mocks base method.
func (m *MockRegisterOrderStorage) SaveOrder(arg0 context.Context, arg1 *entity.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveOrder", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveOrder indicates an expected call of SaveOrder.
func (mr *MockRegisterOrderStorageMockRecorder) SaveOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveOrder", reflect.TypeOf((*MockRegisterOrderStorage)(nil).SaveOrder), arg0, arg1)
}

// UpdateOrderStatusByNumber mocks base method.
func (m *MockRegisterOrderStorage) UpdateOrderStatusByNumber(arg0 context.Context, arg1 int, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrderStatusByNumber", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrderStatusByNumber indicates an expected call of UpdateOrderStatusByNumber.
func (mr *MockRegisterOrderStorageMockRecorder) UpdateOrderStatusByNumber(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrderStatusByNumber", reflect.TypeOf((*MockRegisterOrderStorage)(nil).UpdateOrderStatusByNumber), arg0, arg1, arg2)
}

// UpdateUserBalanceByID mocks base method.
func (m *MockRegisterOrderStorage) UpdateUserBalanceByID(arg0 context.Context, arg1 uint, arg2 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserBalanceByID", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserBalanceByID indicates an expected call of UpdateUserBalanceByID.
func (mr *MockRegisterOrderStorageMockRecorder) UpdateUserBalanceByID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserBalanceByID", reflect.TypeOf((*MockRegisterOrderStorage)(nil).UpdateUserBalanceByID), arg0, arg1, arg2)
}

// MockOrderProcessor is a mock of OrderProcessor interface.
type MockOrderProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockOrderProcessorMockRecorder
}

// MockOrderProcessorMockRecorder is the mock recorder for MockOrderProcessor.
type MockOrderProcessorMockRecorder struct {
	mock *MockOrderProcessor
}

// NewMockOrderProcessor creates a new mock instance.
func NewMockOrderProcessor(ctrl *gomock.Controller) *MockOrderProcessor {
	mock := &MockOrderProcessor{ctrl: ctrl}
	mock.recorder = &MockOrderProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrderProcessor) EXPECT() *MockOrderProcessorMockRecorder {
	return m.recorder
}

// ProcessOrder mocks base method.
func (m *MockOrderProcessor) ProcessOrder(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ProcessOrder", arg0)
}

// ProcessOrder indicates an expected call of ProcessOrder.
func (mr *MockOrderProcessorMockRecorder) ProcessOrder(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessOrder", reflect.TypeOf((*MockOrderProcessor)(nil).ProcessOrder), arg0)
}

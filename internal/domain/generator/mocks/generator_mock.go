package mocks

import (
	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
	"github.com/stretchr/testify/mock"
)

// MockGenerator is a mock type for the Generator type
type MockGenerator struct {
	mock.Mock
}

// Generate provides a mock function with given fields: root, packageName, funcName
func (_m *MockGenerator) Generate(root ast.Expr, packageName string, funcName string) (string, error) {
	ret := _m.Called(root, packageName, funcName)

	var r0 string
	if rf, ok := ret.Get(0).(func(ast.Expr, string, string) string); ok {
		r0 = rf(root, packageName, funcName)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ast.Expr, string, string) error); ok {
		r1 = rf(root, packageName, funcName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockGenerator creates a new instance of MockGenerator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockGenerator(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGenerator {
	mock := &MockGenerator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

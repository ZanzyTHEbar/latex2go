package mocks

import (
	"github.com/ZanzyTHEbar/latex2go/internal/domain/ast"
	"github.com/stretchr/testify/mock"
)

// MockParser is a mock type for the Parser type
type MockParser struct {
	mock.Mock
}

// Parse provides a mock function with given fields: latexString
func (_m *MockParser) Parse(latexString string) (ast.Expr, error) {
	ret := _m.Called(latexString)

	var r0 ast.Expr
	if rf, ok := ret.Get(0).(func(string) ast.Expr); ok {
		r0 = rf(latexString)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ast.Expr)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(latexString)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockParser creates a new instance of MockParser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockParser(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockParser {
	mock := &MockParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

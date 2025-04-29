package mocks

import (
	"github.com/ZanzyTHEbar/latex2go/internal/app"
	"github.com/stretchr/testify/mock"
)

// MockLatexProvider is a mock type for the LatexProvider type
type MockLatexProvider struct {
	mock.Mock
}

// GetLatexInput provides a mock function with given fields:
func (_m *MockLatexProvider) GetLatexInput() (string, app.Config, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 app.Config
	if rf, ok := ret.Get(1).(func() app.Config); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(app.Config)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewMockLatexProvider creates a new instance of MockLatexProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockLatexProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLatexProvider {
	mock := &MockLatexProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockGoCodeWriter is a mock type for the GoCodeWriter type
type MockGoCodeWriter struct {
	mock.Mock
}

// WriteGoCode provides a mock function with given fields: code
func (_m *MockGoCodeWriter) WriteGoCode(code string) error {
	ret := _m.Called(code)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(code)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockGoCodeWriter creates a new instance of MockGoCodeWriter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockGoCodeWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGoCodeWriter {
	mock := &MockGoCodeWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

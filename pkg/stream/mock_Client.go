// Code generated by mockery v1.0.0. DO NOT EDIT.

package stream

import mock "github.com/stretchr/testify/mock"

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

// Add provides a mock function with given fields: key, value
func (_m *MockClient) Add(key string, value string) error {
	ret := _m.Called(key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: key
func (_m *MockClient) Delete(key string) error {
	ret := _m.Called(key)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: key, id
func (_m *MockClient) Get(key string, id string) (string, string, error) {
	ret := _m.Called(key, id)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(key, id)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(string, string) string); ok {
		r1 = rf(key, id)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, string) error); ok {
		r2 = rf(key, id)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

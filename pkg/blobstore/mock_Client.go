// Code generated by mockery v1.0.0. DO NOT EDIT.

package blobstore

import io "io"
import mock "github.com/stretchr/testify/mock"

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

// Delete provides a mock function with given fields: path
func (_m *MockClient) Delete(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: path
func (_m *MockClient) Get(path string) (io.Reader, error) {
	ret := _m.Called(path)

	var r0 io.Reader
	if rf, ok := ret.Get(0).(func(string) io.Reader); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Reader)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsNotExist provides a mock function with given fields: _a0
func (_m *MockClient) IsNotExist(_a0 error) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(error) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Put provides a mock function with given fields: path, reader, objectSize
func (_m *MockClient) Put(path string, reader io.Reader, objectSize int64) error {
	ret := _m.Called(path, reader, objectSize)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, io.Reader, int64) error); ok {
		r0 = rf(path, reader, objectSize)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

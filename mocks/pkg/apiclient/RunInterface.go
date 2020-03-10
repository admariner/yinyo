// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import apiclient "github.com/openaustralia/yinyo/pkg/apiclient"
import io "io"
import mock "github.com/stretchr/testify/mock"
import protocol "github.com/openaustralia/yinyo/pkg/protocol"

// RunInterface is an autogenerated mock type for the RunInterface type
type RunInterface struct {
	mock.Mock
}

// CreateEvent provides a mock function with given fields: event
func (_m *RunInterface) CreateEvent(event protocol.Event) (int, error) {
	ret := _m.Called(event)

	var r0 int
	if rf, ok := ret.Get(0).(func(protocol.Event) int); ok {
		r0 = rf(event)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(protocol.Event) error); ok {
		r1 = rf(event)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateFinishEvent provides a mock function with given fields: stage, exitData
func (_m *RunInterface) CreateFinishEvent(stage string, exitData protocol.ExitDataStage) (int, error) {
	ret := _m.Called(stage, exitData)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, protocol.ExitDataStage) int); ok {
		r0 = rf(stage, exitData)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, protocol.ExitDataStage) error); ok {
		r1 = rf(stage, exitData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateLastEvent provides a mock function with given fields:
func (_m *RunInterface) CreateLastEvent() (int, error) {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateLogEvent provides a mock function with given fields: stage, stream, text
func (_m *RunInterface) CreateLogEvent(stage string, stream string, text string) (int, error) {
	ret := _m.Called(stage, stream, text)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, string, string) int); ok {
		r0 = rf(stage, stream, text)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(stage, stream, text)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateNetworkEvent provides a mock function with given fields: in, out
func (_m *RunInterface) CreateNetworkEvent(in uint64, out uint64) (int, error) {
	ret := _m.Called(in, out)

	var r0 int
	if rf, ok := ret.Get(0).(func(uint64, uint64) int); ok {
		r0 = rf(in, out)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64, uint64) error); ok {
		r1 = rf(in, out)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateStartEvent provides a mock function with given fields: stage
func (_m *RunInterface) CreateStartEvent(stage string) (int, error) {
	ret := _m.Called(stage)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(stage)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(stage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields:
func (_m *RunInterface) Delete() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetApp provides a mock function with given fields:
func (_m *RunInterface) GetApp() (io.ReadCloser, error) {
	ret := _m.Called()

	var r0 io.ReadCloser
	if rf, ok := ret.Get(0).(func() io.ReadCloser); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAppToDirectory provides a mock function with given fields: dir
func (_m *RunInterface) GetAppToDirectory(dir string) error {
	ret := _m.Called(dir)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(dir)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetCache provides a mock function with given fields:
func (_m *RunInterface) GetCache() (io.ReadCloser, error) {
	ret := _m.Called()

	var r0 io.ReadCloser
	if rf, ok := ret.Get(0).(func() io.ReadCloser); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCacheToDirectory provides a mock function with given fields: dir
func (_m *RunInterface) GetCacheToDirectory(dir string) error {
	ret := _m.Called(dir)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(dir)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetCacheToFile provides a mock function with given fields: path
func (_m *RunInterface) GetCacheToFile(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetEvents provides a mock function with given fields: lastID
func (_m *RunInterface) GetEvents(lastID string) (*apiclient.EventIterator, error) {
	ret := _m.Called(lastID)

	var r0 *apiclient.EventIterator
	if rf, ok := ret.Get(0).(func(string) *apiclient.EventIterator); ok {
		r0 = rf(lastID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apiclient.EventIterator)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(lastID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetExitData provides a mock function with given fields:
func (_m *RunInterface) GetExitData() (protocol.ExitData, error) {
	ret := _m.Called()

	var r0 protocol.ExitData
	if rf, ok := ret.Get(0).(func() protocol.ExitData); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(protocol.ExitData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetID provides a mock function with given fields:
func (_m *RunInterface) GetID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetOutput provides a mock function with given fields:
func (_m *RunInterface) GetOutput() (io.ReadCloser, error) {
	ret := _m.Called()

	var r0 io.ReadCloser
	if rf, ok := ret.Get(0).(func() io.ReadCloser); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutputToFile provides a mock function with given fields: path
func (_m *RunInterface) GetOutputToFile(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutApp provides a mock function with given fields: data
func (_m *RunInterface) PutApp(data io.Reader) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func(io.Reader) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutAppFromDirectory provides a mock function with given fields: dir, ignorePaths
func (_m *RunInterface) PutAppFromDirectory(dir string, ignorePaths []string) error {
	ret := _m.Called(dir, ignorePaths)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []string) error); ok {
		r0 = rf(dir, ignorePaths)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutCache provides a mock function with given fields: data
func (_m *RunInterface) PutCache(data io.Reader) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func(io.Reader) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutCacheFromDirectory provides a mock function with given fields: dir
func (_m *RunInterface) PutCacheFromDirectory(dir string) error {
	ret := _m.Called(dir)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(dir)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutOutput provides a mock function with given fields: data
func (_m *RunInterface) PutOutput(data io.Reader) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func(io.Reader) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutOutputFromFile provides a mock function with given fields: path
func (_m *RunInterface) PutOutputFromFile(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields: options
func (_m *RunInterface) Start(options *protocol.StartRunOptions) error {
	ret := _m.Called(options)

	var r0 error
	if rf, ok := ret.Get(0).(func(*protocol.StartRunOptions) error); ok {
		r0 = rf(options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Code generated by mockery v2.43.0. DO NOT EDIT.

package translator

import mock "github.com/stretchr/testify/mock"

// Mock_Downloader is an autogenerated mock type for the Downloader type
type Mock_Downloader struct {
	mock.Mock
}

type Mock_Downloader_Expecter struct {
	mock *mock.Mock
}

func (_m *Mock_Downloader) EXPECT() *Mock_Downloader_Expecter {
	return &Mock_Downloader_Expecter{mock: &_m.Mock}
}

// Download provides a mock function with given fields: url
func (_m *Mock_Downloader) Download(url string) (string, error) {
	ret := _m.Called(url)

	if len(ret) == 0 {
		panic("no return value specified for Download")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(url)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(url)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mock_Downloader_Download_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Download'
type Mock_Downloader_Download_Call struct {
	*mock.Call
}

// Download is a helper method to define mock.On call
//   - url string
func (_e *Mock_Downloader_Expecter) Download(url interface{}) *Mock_Downloader_Download_Call {
	return &Mock_Downloader_Download_Call{Call: _e.mock.On("Download", url)}
}

func (_c *Mock_Downloader_Download_Call) Run(run func(url string)) *Mock_Downloader_Download_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Mock_Downloader_Download_Call) Return(_a0 string, _a1 error) *Mock_Downloader_Download_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Mock_Downloader_Download_Call) RunAndReturn(run func(string) (string, error)) *Mock_Downloader_Download_Call {
	_c.Call.Return(run)
	return _c
}

// NewMock_Downloader creates a new instance of Mock_Downloader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMock_Downloader(t interface {
	mock.TestingT
	Cleanup(func())
}) *Mock_Downloader {
	mock := &Mock_Downloader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

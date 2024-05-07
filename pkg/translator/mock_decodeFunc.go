// Code generated by mockery v2.43.0. DO NOT EDIT.

package translator

import mock "github.com/stretchr/testify/mock"

// Mock_decodeFunc is an autogenerated mock type for the decodeFunc type
type Mock_decodeFunc struct {
	mock.Mock
}

type Mock_decodeFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *Mock_decodeFunc) EXPECT() *Mock_decodeFunc_Expecter {
	return &Mock_decodeFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: data, args
func (_m *Mock_decodeFunc) Execute(data []byte, args *Lingualeo) error {
	ret := _m.Called(data, args)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, *Lingualeo) error); ok {
		r0 = rf(data, args)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Mock_decodeFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type Mock_decodeFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - data []byte
//   - args *Lingualeo
func (_e *Mock_decodeFunc_Expecter) Execute(data interface{}, args interface{}) *Mock_decodeFunc_Execute_Call {
	return &Mock_decodeFunc_Execute_Call{Call: _e.mock.On("Execute", data, args)}
}

func (_c *Mock_decodeFunc_Execute_Call) Run(run func(data []byte, args *Lingualeo)) *Mock_decodeFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte), args[1].(*Lingualeo))
	})
	return _c
}

func (_c *Mock_decodeFunc_Execute_Call) Return(_a0 error) *Mock_decodeFunc_Execute_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Mock_decodeFunc_Execute_Call) RunAndReturn(run func([]byte, *Lingualeo) error) *Mock_decodeFunc_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewMock_decodeFunc creates a new instance of Mock_decodeFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMock_decodeFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *Mock_decodeFunc {
	mock := &Mock_decodeFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
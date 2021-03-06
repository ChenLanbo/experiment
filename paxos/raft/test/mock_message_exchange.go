// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/chenlanbo/experiment/paxos/raft/rpc (interfaces: MessageExchange)

package test

import (
	protos "github.com/chenlanbo/experiment/paxos/protos"
	gomock "github.com/golang/mock/gomock"
)

// Mock of MessageExchange interface
type MockMessageExchange struct {
	ctrl     *gomock.Controller
	recorder *_MockMessageExchangeRecorder
}

// Recorder for MockMessageExchange (not exported)
type _MockMessageExchangeRecorder struct {
	mock *MockMessageExchange
}

func NewMockMessageExchange(ctrl *gomock.Controller) *MockMessageExchange {
	mock := &MockMessageExchange{ctrl: ctrl}
	mock.recorder = &_MockMessageExchangeRecorder{mock}
	return mock
}

func (_m *MockMessageExchange) EXPECT() *_MockMessageExchangeRecorder {
	return _m.recorder
}

func (_m *MockMessageExchange) Append(_param0 string, _param1 *protos.AppendRequest) (*protos.AppendReply, error) {
	ret := _m.ctrl.Call(_m, "Append", _param0, _param1)
	ret0, _ := ret[0].(*protos.AppendReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMessageExchangeRecorder) Append(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Append", arg0, arg1)
}

func (_m *MockMessageExchange) Get(_param0 string, _param1 *protos.GetRequest) (*protos.GetReply, error) {
	ret := _m.ctrl.Call(_m, "Get", _param0, _param1)
	ret0, _ := ret[0].(*protos.GetReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMessageExchangeRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0, arg1)
}

func (_m *MockMessageExchange) Put(_param0 string, _param1 *protos.PutRequest) (*protos.PutReply, error) {
	ret := _m.ctrl.Call(_m, "Put", _param0, _param1)
	ret0, _ := ret[0].(*protos.PutReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMessageExchangeRecorder) Put(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Put", arg0, arg1)
}

func (_m *MockMessageExchange) Vote(_param0 string, _param1 *protos.VoteRequest) (*protos.VoteReply, error) {
	ret := _m.ctrl.Call(_m, "Vote", _param0, _param1)
	ret0, _ := ret[0].(*protos.VoteReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMessageExchangeRecorder) Vote(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Vote", arg0, arg1)
}

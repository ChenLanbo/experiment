// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/chenlanbo/experiment/paxos/protos (interfaces: RaftServer)

package test

import (
	context "golang.org/x/net/context"
	protos "github.com/chenlanbo/experiment/paxos/protos"
	gomock "github.com/golang/mock/gomock"
)

// Mock of RaftServer interface
type MockRaftServer struct {
	ctrl     *gomock.Controller
	recorder *_MockRaftServerRecorder
}

// Recorder for MockRaftServer (not exported)
type _MockRaftServerRecorder struct {
	mock *MockRaftServer
}

func NewMockRaftServer(ctrl *gomock.Controller) *MockRaftServer {
	mock := &MockRaftServer{ctrl: ctrl}
	mock.recorder = &_MockRaftServerRecorder{mock}
	return mock
}

func (_m *MockRaftServer) EXPECT() *_MockRaftServerRecorder {
	return _m.recorder
}

func (_m *MockRaftServer) Append(_param0 context.Context, _param1 *protos.AppendRequest) (*protos.AppendReply, error) {
	ret := _m.ctrl.Call(_m, "Append", _param0, _param1)
	ret0, _ := ret[0].(*protos.AppendReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRaftServerRecorder) Append(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Append", arg0, arg1)
}

func (_m *MockRaftServer) Get(_param0 context.Context, _param1 *protos.GetRequest) (*protos.GetReply, error) {
	ret := _m.ctrl.Call(_m, "Get", _param0, _param1)
	ret0, _ := ret[0].(*protos.GetReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRaftServerRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0, arg1)
}

func (_m *MockRaftServer) Put(_param0 context.Context, _param1 *protos.PutRequest) (*protos.PutReply, error) {
	ret := _m.ctrl.Call(_m, "Put", _param0, _param1)
	ret0, _ := ret[0].(*protos.PutReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRaftServerRecorder) Put(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Put", arg0, arg1)
}

func (_m *MockRaftServer) Vote(_param0 context.Context, _param1 *protos.VoteRequest) (*protos.VoteReply, error) {
	ret := _m.ctrl.Call(_m, "Vote", _param0, _param1)
	ret0, _ := ret[0].(*protos.VoteReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockRaftServerRecorder) Vote(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Vote", arg0, arg1)
}

// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Apache License 2.0 which is available at
// https://www.apache.org/licenses/LICENSE-2.0
//
// SPDX-License-Identifier: Apache-2.0
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/eclipse/paho.mqtt.golang/client.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	"github.com/eclipse/paho.mqtt.golang"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// AddRoute mocks base method.
func (m *MockClient) AddRoute(topic string, callback mqtt.MessageHandler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddRoute", topic, callback)
}

// AddRoute indicates an expected call of AddRoute.
func (mr *MockClientMockRecorder) AddRoute(topic, callback interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddRoute", reflect.TypeOf((*MockClient)(nil).AddRoute), topic, callback)
}

// Connect mocks base method.
func (m *MockClient) Connect() mqtt.Token {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect")
	ret0, _ := ret[0].(mqtt.Token)
	return ret0
}

// Connect indicates an expected call of Connect.
func (mr *MockClientMockRecorder) Connect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockClient)(nil).Connect))
}

// Disconnect mocks base method.
func (m *MockClient) Disconnect(quiesce uint) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Disconnect", quiesce)
}

// Disconnect indicates an expected call of Disconnect.
func (mr *MockClientMockRecorder) Disconnect(quiesce interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Disconnect", reflect.TypeOf((*MockClient)(nil).Disconnect), quiesce)
}

// IsConnected mocks base method.
func (m *MockClient) IsConnected() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsConnected")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsConnected indicates an expected call of IsConnected.
func (mr *MockClientMockRecorder) IsConnected() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsConnected", reflect.TypeOf((*MockClient)(nil).IsConnected))
}

// IsConnectionOpen mocks base method.
func (m *MockClient) IsConnectionOpen() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsConnectionOpen")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsConnectionOpen indicates an expected call of IsConnectionOpen.
func (mr *MockClientMockRecorder) IsConnectionOpen() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsConnectionOpen", reflect.TypeOf((*MockClient)(nil).IsConnectionOpen))
}

// OptionsReader mocks base method.
func (m *MockClient) OptionsReader() mqtt.ClientOptionsReader {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OptionsReader")
	ret0, _ := ret[0].(mqtt.ClientOptionsReader)
	return ret0
}

// OptionsReader indicates an expected call of OptionsReader.
func (mr *MockClientMockRecorder) OptionsReader() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OptionsReader", reflect.TypeOf((*MockClient)(nil).OptionsReader))
}

// Publish mocks base method.
func (m *MockClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", topic, qos, retained, payload)
	ret0, _ := ret[0].(mqtt.Token)
	return ret0
}

// Publish indicates an expected call of Publish.
func (mr *MockClientMockRecorder) Publish(topic, qos, retained, payload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockClient)(nil).Publish), topic, qos, retained, payload)
}

// Subscribe mocks base method.
func (m *MockClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", topic, qos, callback)
	ret0, _ := ret[0].(mqtt.Token)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockClientMockRecorder) Subscribe(topic, qos, callback interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockClient)(nil).Subscribe), topic, qos, callback)
}

// SubscribeMultiple mocks base method.
func (m *MockClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribeMultiple", filters, callback)
	ret0, _ := ret[0].(mqtt.Token)
	return ret0
}

// SubscribeMultiple indicates an expected call of SubscribeMultiple.
func (mr *MockClientMockRecorder) SubscribeMultiple(filters, callback interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeMultiple", reflect.TypeOf((*MockClient)(nil).SubscribeMultiple), filters, callback)
}

// Unsubscribe mocks base method.
func (m *MockClient) Unsubscribe(topics ...string) mqtt.Token {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range topics {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Unsubscribe", varargs...)
	ret0, _ := ret[0].(mqtt.Token)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe.
func (mr *MockClientMockRecorder) Unsubscribe(topics ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockClient)(nil).Unsubscribe), topics...)
}

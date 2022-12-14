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
// Source: github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration (interfaces: UpdateManager)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MockUpdateManager is a mock of UpdateManager interface.
type MockUpdateManager struct {
	ctrl     *gomock.Controller
	recorder *MockUpdateManagerMockRecorder
}

// MockUpdateManagerMockRecorder is the mock recorder for MockUpdateManager.
type MockUpdateManagerMockRecorder struct {
	mock *MockUpdateManager
}

// NewMockUpdateManager creates a new mock instance.
func NewMockUpdateManager(ctrl *gomock.Controller) *MockUpdateManager {
	mock := &MockUpdateManager{ctrl: ctrl}
	mock.recorder = &MockUpdateManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUpdateManager) EXPECT() *MockUpdateManagerMockRecorder {
	return m.recorder
}

// Apply mocks base method.
func (m *MockUpdateManager) Apply(ctx context.Context, mf []*unstructured.Unstructured) interface{} {
	m.ctrl.T.Helper()
	return m.ctrl.Call(m, "Apply", ctx, mf)[0]
}

// Apply indicates an expected call of Apply.
func (mr *MockUpdateManagerMockRecorder) Apply(ctx, mf interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Apply", reflect.TypeOf((*MockUpdateManager)(nil).Apply), ctx, mf)
}

// Dispose mocks base method.
func (m *MockUpdateManager) Dispose(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dispose", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Dispose indicates an expected call of Dispose.
func (mr *MockUpdateManagerMockRecorder) Dispose(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dispose", reflect.TypeOf((*MockUpdateManager)(nil).Dispose), ctx)
}

// Get mocks base method.
func (m *MockUpdateManager) Get(ctx context.Context) []*unstructured.Unstructured {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx)
	ret0, _ := ret[0].([]*unstructured.Unstructured)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockUpdateManagerMockRecorder) Get(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockUpdateManager)(nil).Get), ctx)
}

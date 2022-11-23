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

package util

import (
	"errors"
	"testing"
	"time"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	mocksmqtt "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/mqtt"
	"github.com/golang/mock/gomock"
)

func TestMqttConnect(t *testing.T) {
	backoffInterval = 500 * time.Millisecond
	tests := map[string]struct {
		mockExecution func(t *testing.T, controller *gomock.Controller, mockClient *mocksmqtt.MockClient)
	}{
		"test_connect_from_first_attempt": {
			mockExecution: func(t *testing.T, controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				c := make(chan struct{}, 1)
				c <- struct{}{}

				mockTokenSuccess := mocksmqtt.NewMockToken(controller)
				mockClient.EXPECT().Connect().Return(mockTokenSuccess)
				mockTokenSuccess.EXPECT().Done().Return(c)
				mockTokenSuccess.EXPECT().Error().Return(nil)
			},
		},
		"test_connect_from_next_attemppt": {
			mockExecution: func(t *testing.T, controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				c := make(chan struct{}, 3)
				c <- struct{}{}
				c <- struct{}{}
				c <- struct{}{}
				mockTokenError := mocksmqtt.NewMockToken(controller)
				mockTokenSuccess := mocksmqtt.NewMockToken(controller)

				mockClient.EXPECT().Connect().Return(mockTokenError)
				mockTokenError.EXPECT().Done().Return(c)
				mockTokenError.EXPECT().Error().Return(errors.New("first try unsuccessful"))

				mockClient.EXPECT().Connect().Return(mockTokenError)
				mockTokenError.EXPECT().Done().Return(c)
				mockTokenError.EXPECT().Error().Return(errors.New("second try unsuccessful"))

				mockClient.EXPECT().Connect().Return(mockTokenSuccess)
				mockTokenSuccess.EXPECT().Done().Return(c)
				mockTokenSuccess.EXPECT().Error().Return(nil)
			},
		},
	}

	controller := gomock.NewController(t)
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			mockClient := mocksmqtt.NewMockClient(controller)
			testCase.mockExecution(t, controller, mockClient)
			testutil.AssertNil(t, MqttConnect(mockClient, "test-mqtt:1883"))
		})
	}
}

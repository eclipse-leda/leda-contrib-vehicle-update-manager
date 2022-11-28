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

package selfupdate

import (
	mocksmqtt "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/mock/gomock"
)

func setupMockClient(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
	mockSubToken := setupMockPubSubToken(controller)
	mockPubToken := setupMockPubSubToken(controller)
	mockClient.EXPECT().Subscribe(topicSelfUpdateDesiredStateFeedback, gomock.Any(), gomock.Any()).Return(mockSubToken)
	mockClient.EXPECT().Unsubscribe(topicSelfUpdateDesiredStateFeedback)
	mockClient.EXPECT().Publish(topicSelfUpdateDesiredState, gomock.Any(), gomock.Any(), gomock.Any()).Return(mockPubToken)
}

func setupMockPubSubToken(controller *gomock.Controller) mqtt.Token {
	mockPubSubToken := mocksmqtt.NewMockToken(controller)
	mockPubSubToken.EXPECT().WaitTimeout(gomock.Any()).Return(true)
	return mockPubSubToken
}

func setupMockMessage(controller *gomock.Controller, payload string) mqtt.Message {
	mockMessage := mocksmqtt.NewMockMessage(controller)
	mockMessage.EXPECT().Payload().Return([]byte(payload))
	return mockMessage
}

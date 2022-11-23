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

package updateorchestrator

import (
	"context"
	"testing"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	mocksevents "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/events"
	mocksorchmgr "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/orchestration"

	mocksmqtt "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/mock/gomock"
)

func TestSubscribeRemoteConnectionStatusSuccessful(t *testing.T) {
	controller := gomock.NewController(t)

	mockClient := mocksmqtt.NewMockClient(controller)
	mockSubToken := mocksmqtt.NewMockToken(controller)
	mockSubToken.EXPECT().WaitTimeout(gomock.Any()).Return(true)
	mockClient.EXPECT().Subscribe(topicRemoteStatus, gomock.Any(), gomock.Any()).Return(mockSubToken)

	updOrc := &updateOrchestrator{
		pahoClient: mockClient,
		cfg:        &mgrOpts{},
	}

	testutil.AssertNil(t, updOrc.subscribeRemoteConnectionStatus())
}
func TestSubscribeRemoteConnectionStatusTimeoutErr(t *testing.T) {
	controller := gomock.NewController(t)

	mockSubToken := mocksmqtt.NewMockToken(controller)
	mockSubToken.EXPECT().WaitTimeout(gomock.Any()).Return(false)

	mockClient := mocksmqtt.NewMockClient(controller)
	mockClient.EXPECT().Subscribe(topicRemoteStatus, gomock.Any(), gomock.Any()).Return(mockSubToken)

	updOrc := &updateOrchestrator{
		pahoClient: mockClient,
		cfg:        &mgrOpts{},
	}

	testutil.AssertNotNil(t, updOrc.subscribeRemoteConnectionStatus())
}

func TestHandleConnectionStatus(t *testing.T) {
	type mockFunc func(*mocksevents.MockUpdateEventsManager, *mocksorchmgr.MockUpdateManager, *mocksorchmgr.MockUpdateManager)

	tests := []struct {
		name                    string
		payload                 string
		currentStatus           bool
		updatedConnectionStatus bool
		mockExecution           mockFunc
	}{
		{
			name:                    "payload-json-empty",
			payload:                 "",
			currentStatus:           false,
			updatedConnectionStatus: false,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {
			},
		},
		{
			name:                    "payload-wrong-format",
			payload:                 "wrong-json-format",
			currentStatus:           false,
			updatedConnectionStatus: false,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {
			},
		},
		{
			name: "payload-connected-key-missing",
			payload: `{
					"connectedKeyMissing": false
				}
				`,
			currentStatus:           false,
			updatedConnectionStatus: false,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {
			},
		},
		{
			name: "payload-connected-key-wrong-format",
			payload: `{
					"connected": 1
				}
				`,
			currentStatus:           false,
			updatedConnectionStatus: false,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {
			},
		},
		{
			name: "connection-status-key-connected-false-current-true",
			payload: `{
				"connected": false
			}
			`,
			currentStatus:           true,
			updatedConnectionStatus: false,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {
			},
		},
		{
			name: "connection-status-key-connected-true-current-false",
			payload: `{
				"connected": true
			}
			`,
			currentStatus:           false,
			updatedConnectionStatus: true,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {

				mockSelfUpdateMgr.EXPECT().Get(context.Background())
				mockK8sOrchestrationMgr.EXPECT().Get(context.Background())
				mockEventsMgr.EXPECT().Publish(context.Background(), gomock.Any()).Do(func(ctx context.Context, event *events.Event) {
					testutil.AssertEqual(t, events.EventActionResourcesUpdated, event.Action)
					testutil.AssertEqual(t, events.EventTypeResources, event.Type)
					testutil.AssertEqual(t, ctx, event.Context)
					testutil.AssertEqual(t, nil, event.Error)
				})
			},
		},
		{
			name: "connection-status-key-connected-true-current-true",
			payload: `{
				"connected": true
			}
			`,
			currentStatus:           true,
			updatedConnectionStatus: true,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {
			},
		},
		{
			name: "connection-status-key-connected-false-current-false",
			payload: `{
				"connected": false
			}
			`,
			currentStatus:           false,
			updatedConnectionStatus: false,
			mockExecution: func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager) {
			},
		},
	}
	controller := gomock.NewController(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name)

			mockMessage := setupMockMessage(controller, test.payload)
			mockClient := mocksmqtt.NewMockClient(controller)

			selfUpdateMgr := mocksorchmgr.NewMockUpdateManager(controller)
			k8sUpdateMgr := mocksorchmgr.NewMockUpdateManager(controller)
			eventsMgr := mocksevents.NewMockUpdateEventsManager(controller)

			test.mockExecution(eventsMgr, selfUpdateMgr, k8sUpdateMgr)

			updOrc := &updateOrchestrator{
				pahoClient:              mockClient,
				selfUpdateManager:       selfUpdateMgr,
				k8sOrchestrationManager: k8sUpdateMgr,
				eventsManager:           eventsMgr,
				connectionStatus:        test.currentStatus,
			}
			updOrc.handleConnectionStatus(mockClient, mockMessage)

			testutil.AssertEqual(t, test.updatedConnectionStatus, updOrc.connectionStatus)
		})
	}
}

func setupMockMessage(controller *gomock.Controller, payload string) mqtt.Message {
	mockMessage := mocksmqtt.NewMockMessage(controller)
	mockMessage.EXPECT().Payload().Return([]byte(payload)).Times(2)
	return mockMessage
}

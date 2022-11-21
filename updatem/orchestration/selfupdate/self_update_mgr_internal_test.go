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
	"reflect"
	"testing"

	mocksevents "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/events"
	mocksmqtt "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/mqtt"
	"github.com/golang/mock/gomock"
)

func TestSubscribeSelfUpdateCurrentState(t *testing.T) {
	controller := gomock.NewController(t)
	tests := []struct {
		name       string
		timeoutErr bool
	}{
		{
			name:       "subscribe_timeout",
			timeoutErr: true,
		},
		{
			name:       "subscribe_succesfully",
			timeoutErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubToken := mocksmqtt.NewMockToken(controller)
			mockSubToken.EXPECT().WaitTimeout(gomock.Any()).Return(!tt.timeoutErr)
			mockClient := mocksmqtt.NewMockClient(controller)

			mockClient.EXPECT().Subscribe(topicSelfUpdateCurrentState, gomock.Any(), gomock.Any()).Return(mockSubToken)

			selfUpdateMgr := &selfUpdateManager{
				pahoClient: mockClient,
				cfg: &mgrOpts{
					acknowledgeTimeout: 10000,
				},
			}
			if err := selfUpdateMgr.subscribeSelfUpdateCurrentState(); (err != nil) != tt.timeoutErr {
				t.Fail()
			}
		})
	}
}

func TestSelfUpdateCurrentState(t *testing.T) {
	selfUpdateCurrentState := `
apiVersion: "sdv.eclipse.org/v1"
kind: SelfUpdateBundle
metadata:
  name: self-update-bundle-example
spec:
  bundleVersion: v1beta3
`
	controller := gomock.NewController(t)

	mockClient := mocksmqtt.NewMockClient(controller)
	mockEventsMgr := mocksevents.NewMockUpdateEventsManager(controller)
	mockEventsMgr.EXPECT().Publish(gomock.Any(), gomock.Any())

	selfUpdateMgr := &selfUpdateManager{
		pahoClient: mockClient,
		eventsMgr:  mockEventsMgr,
		cfg: &mgrOpts{
			acknowledgeTimeout: 10000,
		},
	}

	selfUpdateMgr.handleSelfUpdateCurrentState(mockClient, setupMockMessage(controller, selfUpdateCurrentState))

	if selfUpdateMgr.currentState == nil {
		t.Fail()
	}
	_, u, _ := parseMultiYAML([]byte(selfUpdateCurrentState))

	if !reflect.DeepEqual(selfUpdateMgr.currentState, u[0]) {
		t.Fail()
	}
}

func TestSelfUpdateCurrentStateInvalidPayload(t *testing.T) {
	controller := gomock.NewController(t)

	mockClient := mocksmqtt.NewMockClient(controller)
	mockEventsMgr := mocksevents.NewMockUpdateEventsManager(controller)

	selfUpdateMgr := &selfUpdateManager{
		pahoClient: mockClient,
		eventsMgr:  mockEventsMgr,
		cfg: &mgrOpts{
			acknowledgeTimeout: 10000,
		},
	}

	selfUpdateMgr.handleSelfUpdateCurrentState(selfUpdateMgr.pahoClient, setupMockMessage(controller, ""))
	if selfUpdateMgr.currentState != nil {
		t.Fail()
	}
}

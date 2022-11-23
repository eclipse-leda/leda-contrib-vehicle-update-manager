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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	mocksevents "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/events"
	mocksmqtt "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/mqtt"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const selfUpdateManifest = `
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
  name: self-update-bundle-example-3
spec:
  bundleName: swdv-arm64-build42
  bundleVersion: v1beta3
  bundleDownloadUrl: https://example.com/repository/base/
  bundleTarget: base
`

func TestApply(t *testing.T) {
	type mockFunc func(*gomock.Controller, *mocksmqtt.MockClient)
	type notifyResult func(*selfUpdateOperation)

	tests := []struct {
		testName               string
		expectedResult         OperationResult
		enableReboot           bool
		rebootTimeout          string
		expectedError          bool
		expectedRebootRequired bool
		waitOperationInit      bool
		mockExec               mockFunc
		notifyResult           notifyResult
	}{
		{
			"error_subscribing_self_update_feedback",
			SelfUpdateResultError,
			false,
			"",
			true,
			false,
			false,
			func(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				mockSubToken := mocksmqtt.NewMockToken(controller)
				mockSubToken.EXPECT().WaitTimeout(gomock.Any()).Return(false)
				mockClient.EXPECT().Subscribe(topicSelfUpdateDesiredStateFeedback, gomock.Any(), gomock.Any()).Return(mockSubToken)
			},
			nil,
		},
		{
			"error_publishing_self_update_manifest",
			SelfUpdateResultError,
			false,
			"",
			true,
			false,
			false,
			func(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				mockSubToken := mocksmqtt.NewMockToken(controller)
				mockSubToken.EXPECT().WaitTimeout(gomock.Any()).Return(true)
				mockPubToken := mocksmqtt.NewMockToken(controller)
				mockPubToken.EXPECT().WaitTimeout(gomock.Any()).Return(false)
				mockClient.EXPECT().Subscribe(topicSelfUpdateDesiredStateFeedback, gomock.Any(), gomock.Any()).Return(mockSubToken)
				mockClient.EXPECT().Unsubscribe(topicSelfUpdateDesiredStateFeedback)
				mockClient.EXPECT().Publish(topicSelfUpdateDesiredState, gomock.Any(), gomock.Any(), gomock.Any()).Return(mockPubToken)
			},
			nil,
		},
		{
			"apply_self_update_state_installed",
			SelfUpdateResultInstalled,
			false,
			"",
			false,
			false,
			true,
			func(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				setupMockClient(controller, mockClient)
			},
			func(selfUpdateOperation *selfUpdateOperation) {
				selfUpdateOperation.result = SelfUpdateResultInstalled
				selfUpdateOperation.done <- true
			},
		},
		{
			"apply_self_update_state_installed_reboot_required",
			SelfUpdateResultInstalled,
			true,
			"1m",
			false,
			true,
			true,
			func(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				setupMockClient(controller, mockClient)
			},
			func(selfUpdateOperation *selfUpdateOperation) {
				selfUpdateOperation.result = SelfUpdateResultInstalled
				selfUpdateOperation.done <- true
			},
		},
		{
			"apply_self_update_state_error",
			SelfUpdateResultError,
			false,
			"",
			true,
			false,
			true,
			func(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				setupMockClient(controller, mockClient)
			},
			func(selfUpdateOperation *selfUpdateOperation) {
				selfUpdateOperation.result = SelfUpdateResultError
				selfUpdateOperation.err = fmt.Errorf("error apply self update")
				selfUpdateOperation.done <- true
			},
		},
		{
			"apply_self_update_state_rejected",
			SelfUpdateResultRejected,
			false,
			"",
			false,
			false,
			true,
			func(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				setupMockClient(controller, mockClient)
			},
			func(selfUpdateOperation *selfUpdateOperation) {
				selfUpdateOperation.result = SelfUpdateResultRejected
				selfUpdateOperation.done <- true
			},
		},
		{
			"apply_self_update_state_timeout",
			SelfUpdateResultTimeout,
			false,
			"",
			true,
			false,
			true,
			func(controller *gomock.Controller, mockClient *mocksmqtt.MockClient) {
				setupMockClient(controller, mockClient)
			},
			nil,
		},
	}

	controller := gomock.NewController(t)

	for _, testValues := range tests {
		t.Run(testValues.testName, func(t *testing.T) {
			mockClient := mocksmqtt.NewMockClient(controller)
			mockEventsMgr := mocksevents.NewMockUpdateEventsManager(controller)
			if testValues.mockExec != nil {
				testValues.mockExec(controller, mockClient)
			}

			selfUpdateManager := &selfUpdateManager{
				eventsMgr:  mockEventsMgr,
				pahoClient: mockClient,
				cfg: &mgrOpts{
					timeout:       "500ms",
					enableReboot:  testValues.enableReboot,
					rebootTimeout: testValues.rebootTimeout,
				},
				applyLock: sync.Mutex{},
			}

			resChan := make(chan interface{})
			go performTestSelfUpdate(resChan, selfUpdateManager)
			if testValues.waitOperationInit {
				waitSelfUpdateOperationInitiated(selfUpdateManager)
			}

			if testValues.notifyResult != nil {
				testValues.notifyResult(selfUpdateManager.selfUpdateOperation)
			}

			res := <-resChan
			applyResult := res.(*ApplyResult)
			assertSelfUpdateResult(t, testValues.expectedResult, testValues.rebootTimeout, testValues.expectedRebootRequired, testValues.expectedError, *applyResult)
		})
	}
}

func TestGet(t *testing.T) {
	selfUpdateManager := selfUpdateManager{
		currentState: &unstructured.Unstructured{},
	}

	result := selfUpdateManager.Get(context.Background())

	if len(result) != 1 {
		t.Fail()
	}
	if result[0] != selfUpdateManager.currentState {
		t.Fail()
	}
}

func TestDispose(t *testing.T) {
	controller := gomock.NewController(t)
	mockClient := mocksmqtt.NewMockClient(controller)
	mockClient.EXPECT().Disconnect(gomock.Any())

	selfUpdateManager := selfUpdateManager{
		pahoClient: mockClient,
		cfg: &mgrOpts{
			disconnectTimeout: 200,
		},
	}

	if err := selfUpdateManager.Dispose(context.Background()); err != nil {
		t.Fail()
	}
}

func waitSelfUpdateOperationInitiated(selfUpdateManager *selfUpdateManager) {
	for {
		if selfUpdateManager.selfUpdateOperation != nil {
			break
		}
		<-time.After(100 * time.Millisecond)
	}
}

func performTestSelfUpdate(res chan interface{}, selfUpdateManager *selfUpdateManager) {
	_, unstructured, _ := parseMultiYAML([]byte(selfUpdateManifest))
	applyResult := selfUpdateManager.Apply(context.Background(), unstructured).(*ApplyResult)
	res <- applyResult
}

func assertSelfUpdateResult(t *testing.T, expectedResult OperationResult, expectedRebootTimeout string, expectedRebootRequired bool, hasError bool, actual ApplyResult) {
	if expectedResult != actual.Result {
		t.Fail()
	}
	expectedRebootTimeoutDuration, _ := time.ParseDuration(expectedRebootTimeout)
	if expectedRebootTimeoutDuration != actual.RebootTimeout {
		t.Fail()
	}
	if expectedRebootRequired != actual.RebootRequired {
		t.Fail()
	}
	if hasError && (actual.Err == nil) {
		t.Fail()
	}
	if !hasError && (actual.Err != nil) {
		t.Fail()
	}
}

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

package things

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/golang/mock/gomock"
)

func TestSoftwareUpdatableManifestsHandleEvents(t *testing.T) {
	const (
		testEventsTimeout         = 5 * time.Second
		testCorrelationID         = "test-correlation-id"
		testSoftwareModuleName    = "test-software-module-name"
		testSoftwareModuleVersion = "1.0.0"
	)
	var (
		commonTestOperationStatus = &datatypes.OperationStatus{
			CorrelationID: testCorrelationID,
			SoftwareModule: &datatypes.SoftwareModuleID{
				Name:    testSoftwareModuleName,
				Version: testSoftwareModuleVersion,
			},
		}
		commonEventContext = context.WithValue(context.Background(), contextKeyOperationStatus, &suOperationContextValue{
			correlationID:         testCorrelationID,
			softwareModuleName:    testSoftwareModuleName,
			softwareModuleVersion: testSoftwareModuleVersion,
		})
	)

	controller := gomock.NewController(t)

	setupEventsManagerMock(controller)
	setupUpdateManagerMock(controller)
	setupThingMock(controller)

	testSuMfEvents := newSoftwareUpdatableManifests(mockThing, mockEventsManager, mockUpdateManager)
	testSuMfInternal := testSuMfEvents.(*softwareUpdatableManifests)

	defer func() {
		testSuMfEvents.dispose()
		controller.Finish()
	}()
	eventChan := make(chan *events.Event)
	errorChan := make(chan error)

	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
	testSuMfInternal.handleOrchestrationEvents(context.Background())

	mockPropertyChangedEvent := func(t *testing.T, propertyPath string, expectedValue *datatypes.OperationStatus, wg *sync.WaitGroup) {
		wg.Add(1)
		mockThing.EXPECT().SetFeatureProperty(SoftwareUpdatableManifestsFeatureID, propertyPath, gomock.Any()).
			Do(func(fId, propPath string, value *datatypes.OperationStatus) {
				testutil.AssertEqual(t, expectedValue.CorrelationID, value.CorrelationID)
				testutil.AssertEqual(t, expectedValue.Status, value.Status)
				testutil.AssertEqual(t, expectedValue.Message, value.Message)
				testutil.AssertNotNil(t, value.SoftwareModule)
				testutil.AssertEqual(t, expectedValue.SoftwareModule.Name, value.SoftwareModule.Name)
				testutil.AssertEqual(t, expectedValue.SoftwareModule.Version, value.SoftwareModule.Version)

				wg.Done()

			}).Times(1)
	}

	type mockExec func(t *testing.T) *sync.WaitGroup
	tests := map[string]struct {
		chanEvent     *events.Event
		lastOperation *datatypes.OperationStatus
		mockExecution mockExec
	}{
		"test_things_orchestration_events_started": {
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationStarted,
				Context: commonEventContext,
			},
			lastOperation: commonTestOperationStatus,
			mockExecution: func(t *testing.T) *sync.WaitGroup {
				wg := &sync.WaitGroup{}
				mockPropertyChangedEvent(t, softwareUpdatablePropertyLastOperation, &datatypes.OperationStatus{
					CorrelationID: testCorrelationID,
					SoftwareModule: &datatypes.SoftwareModuleID{
						Name:    testSoftwareModuleName,
						Version: testSoftwareModuleVersion,
					},
					Status: datatypes.Installing,
				}, wg)
				return wg
			},
		},
		"test_things_orchestration_events_finished_success": {
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationFinished,
				Context: commonEventContext,
				Error:   nil,
			},
			lastOperation: commonTestOperationStatus,
			mockExecution: func(t *testing.T) *sync.WaitGroup {
				wg := &sync.WaitGroup{}
				mockPropertyChangedEvent(t, softwareUpdatablePropertyLastOperation, &datatypes.OperationStatus{
					CorrelationID: testCorrelationID,
					SoftwareModule: &datatypes.SoftwareModuleID{
						Name:    testSoftwareModuleName,
						Version: testSoftwareModuleVersion,
					},
					Status: datatypes.Installed,
				}, wg)
				mockPropertyChangedEvent(t, softwareUpdatablePropertyLastOperation, &datatypes.OperationStatus{
					CorrelationID: testCorrelationID,
					SoftwareModule: &datatypes.SoftwareModuleID{
						Name:    testSoftwareModuleName,
						Version: testSoftwareModuleVersion,
					},
					Status: datatypes.FinishedSuccess,
				}, wg)
				return wg
			},
		},
		"test_things_orchestration_events_finished_error": {
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationFinished,
				Context: commonEventContext,
				Error:   log.NewError("test error"),
			},
			lastOperation: commonTestOperationStatus,
			mockExecution: func(t *testing.T) *sync.WaitGroup {
				wg := &sync.WaitGroup{}
				opStatus := &datatypes.OperationStatus{
					CorrelationID: testCorrelationID,
					SoftwareModule: &datatypes.SoftwareModuleID{
						Name:    testSoftwareModuleName,
						Version: testSoftwareModuleVersion,
					},
					Status:  datatypes.FinishedError,
					Message: "test error",
				}
				mockPropertyChangedEvent(t, softwareUpdatablePropertyLastFailedOperation, opStatus, wg)
				mockPropertyChangedEvent(t, softwareUpdatablePropertyLastOperation, opStatus, wg)
				return wg
			},
		},
		"test_things_orchestration_events_irrelevant": {
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationRunning,
				Context: commonEventContext,
			},
			lastOperation: commonTestOperationStatus,
			mockExecution: func(t *testing.T) *sync.WaitGroup {
				mockThing.EXPECT().SetFeatureProperty(SoftwareUpdatableManifestsFeatureID, gomock.Any(), gomock.Any()).Times(0)
				return nil
			},
		},
		"test_things_orchestration_events_no_context": {
			chanEvent: &events.Event{
				Type:   orchestration.EventTypeOrchestration,
				Action: orchestration.EventActionOrchestrationFinished,
			},
			lastOperation: commonTestOperationStatus,
			mockExecution: func(t *testing.T) *sync.WaitGroup {
				mockThing.EXPECT().SetFeatureProperty(SoftwareUpdatableManifestsFeatureID, gomock.Any(), gomock.Any()).Times(0)
				return nil
			},
		},
		"test_things_orchestration_events_irrelevant_context": {
			chanEvent: &events.Event{
				Type:   orchestration.EventTypeOrchestration,
				Action: orchestration.EventActionOrchestrationFinished,
				Context: context.WithValue(context.Background(), contextKeyOperationStatus, &suOperationContextValue{
					correlationID:         "-",
					softwareModuleName:    testSoftwareModuleName,
					softwareModuleVersion: testSoftwareModuleVersion,
				}),
			},
			lastOperation: commonTestOperationStatus,
			mockExecution: func(t *testing.T) *sync.WaitGroup {
				mockThing.EXPECT().SetFeatureProperty(SoftwareUpdatableManifestsFeatureID, gomock.Any(), gomock.Any()).Times(0)
				return nil
			},
		},
		"test_things_container_events_irrelevant": {
			chanEvent: &events.Event{
				Type: events.EventTypeResources,
			},
			lastOperation: commonTestOperationStatus,
			mockExecution: func(t *testing.T) *sync.WaitGroup {
				mockThing.EXPECT().SetFeatureProperty(SoftwareUpdatableManifestsFeatureID, gomock.Any(), gomock.Any()).Times(0)
				return nil
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			testSuMfInternal.status.LastOperation = testCase.lastOperation
			testWg := testCase.mockExecution(t)
			eventChan <- testCase.chanEvent
			if testWg != nil {
				testutil.AssertWithTimeout(t, testWg, testEventsTimeout)
			}
		})
	}
}

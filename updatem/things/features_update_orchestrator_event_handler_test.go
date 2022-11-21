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
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const testEventsTimeout = 5 * time.Second

func TestUpdateOrchestratorHandleOrchestrationEvents(t *testing.T) {
	const (
		testMfName    = "test-manifest"
		testMfVersion = "0.1.1"
	)
	controller := setUpMocks(t)

	testCtrOrchestrator := newUpdateOrchestratorFeature(mockThing, mockEventsManager, mockUpdateManager)
	testManifest := getTestManifest()
	testStatus := &updateOrchestratorFeatureStatus{State: &manifestState{
		Manifest: testManifest,
	}}
	type mockUpdateEventOrchestrator func(t *testing.T, evt *events.Event, testWg *sync.WaitGroup)
	assertStatesEqual := func(t *testing.T, expectedMf []*unstructured.Unstructured, expectedStatus manifestStatus, actualState *manifestState) {
		testutil.AssertEqual(t, expectedMf, actualState.Manifest)
		testutil.AssertEqual(t, expectedStatus, actualState.Status)
		if expectedStatus == manifestStatusFinishedError {
			testutil.AssertNotNil(t, actualState.Error)
			testutil.AssertEqual(t, 500, actualState.Error.Code)
			testutil.AssertNotEqual(t, "", actualState.Error.Message)
		}
	}

	defer func() {
		testCtrOrchestrator.dispose()
		controller.Finish()
	}()
	eventChan := make(chan *events.Event, 1)
	errorChan := make(chan error, 1)

	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
	mockUpdateManager.EXPECT().Get(gomock.Any()).Return(nil).Times(2)
	testCtrOrchestrator.(*updateOrchestratorFeature).handleEvents(context.Background())

	tests := map[string]struct {
		chanEvent     *events.Event
		stat          *updateOrchestratorFeatureStatus
		mockExecution mockUpdateEventOrchestrator
	}{
		"test_things_orchestration_started": {
			stat: nil,
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationStarted,
				Context: orchestration.SetUpdateMgrApplyContext(context.Background(), testManifest),
			},
			mockExecution: func(t *testing.T, evt *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				evt.Context = setApplyCorrelationIDContext(evt.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusState, gomock.Any()).Do(
					func(id, path string, state *manifestState) {
						assertStatesEqual(t, testManifest, manifestStatusStarted, state)
						testutil.AssertEqual(t, testCorrelationID, state.CorrelationID)
						testWg.Done()
					})
			},
		},
		"test_things_orchestration_running": {
			stat: testStatus,
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationRunning,
				Context: orchestration.SetUpdateMgrApplyContext(context.Background(), testStatus.State.Manifest),
			},
			mockExecution: func(t *testing.T, evt *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				evt.Context = setApplyCorrelationIDContext(evt.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusState, gomock.Any()).Do(
					func(id, path string, state *manifestState) {
						testutil.AssertEqual(t, testCorrelationID, state.CorrelationID)
						assertStatesEqual(t, testManifest, manifestStatusRunning, state)
						testWg.Done()
					})
			},
		},
		"test_things_orchestration_finished_ok": {
			stat: testStatus,
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationFinished,
				Context: orchestration.SetUpdateMgrApplyContext(context.Background(), testManifest),
			},
			mockExecution: func(t *testing.T, evt *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(2)
				evt.Context = setApplyCorrelationIDContext(evt.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusState, gomock.Any()).Do(
					func(id, path string, state *manifestState) {
						testutil.AssertEqual(t, testCorrelationID, state.CorrelationID)
						assertStatesEqual(t, testManifest, manifestStatusFinishedSuccess, state)
						testWg.Done()
					})
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Do(
					func(id, path string, unstructured []*unstructured.Unstructured) {
						testutil.AssertNil(t, unstructured)
						testWg.Done()
					})
			},
		},
		"test_things_orchestration_finished_error": {
			stat: testStatus,
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationFinished,
				Context: orchestration.SetUpdateMgrApplyContext(context.Background(), testManifest),
				Error:   log.NewError("test error"),
			},
			mockExecution: func(t *testing.T, evt *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(2)
				evt.Context = setApplyCorrelationIDContext(evt.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusState, gomock.Any()).Do(
					func(id, path string, state *manifestState) {
						testutil.AssertEqual(t, testCorrelationID, state.CorrelationID)
						assertStatesEqual(t, testManifest, manifestStatusFinishedError, state)
						testWg.Done()
					})
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Do(
					func(id, path string, unstructured []*unstructured.Unstructured) {
						testutil.AssertNil(t, unstructured)
						testWg.Done()
					})
			},
		},
		"test_things_orchestration_running_no_cfg": {
			stat: nil,
			chanEvent: &events.Event{
				Type:    orchestration.EventTypeOrchestration,
				Action:  orchestration.EventActionOrchestrationRunning,
				Context: orchestration.SetUpdateMgrApplyContext(context.Background(), testManifest),
			},
			mockExecution: func(t *testing.T, evt *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusState, gomock.Any()).Times(0)
				testWg.Done()
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testWg := &sync.WaitGroup{}
			testCtrOrchestrator.(*updateOrchestratorFeature).status = testCase.stat

			testCase.mockExecution(t, testCase.chanEvent, testWg)
			eventChan <- testCase.chanEvent
			testutil.AssertWithTimeout(t, testWg, testEventsTimeout)
		})
	}
}

func TestUpdateOrchestratorHandleResourceEvents(t *testing.T) {
	defaultCurrentStateDelay = 500 * time.Millisecond

	controller := setUpMocks(t)

	testCtrOrchestrator := newUpdateOrchestratorFeature(mockThing, mockEventsManager, mockUpdateManager)
	type mockUpdateEventOrchestrator func(t *testing.T, ctrEvent *events.Event, testWg *sync.WaitGroup)

	defer func() {
		testCtrOrchestrator.dispose()
		controller.Finish()
	}()
	eventChan := make(chan *events.Event, 1)
	errorChan := make(chan error, 1)

	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Return(eventChan, errorChan)
	mockUpdateManager.EXPECT().Get(gomock.Any()).Return(nil).Times(5)
	testCtrOrchestrator.(*updateOrchestratorFeature).handleEvents(context.Background())

	tests := map[string]struct {
		chanEvent     *events.Event
		mockExecution mockUpdateEventOrchestrator
	}{
		"test_action_resource_added": {
			chanEvent: &events.Event{
				Type:    events.EventTypeResources,
				Action:  events.EventActionResourcesAdded,
				Context: context.Background(),
				Source: &unstructured.Unstructured{
					Object: map[string]interface{}{"kind": "SelfUpdateBundle"},
				},
			},
			mockExecution: func(t *testing.T, ctrEvent *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				ctrEvent.Context = setApplyCorrelationIDContext(ctrEvent.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Do(
					func(id, path string, unstructured []*unstructured.Unstructured) {
						testutil.AssertNil(t, unstructured)
						testWg.Done()
					})
			},
		},
		"test_action_resource_updated": {
			chanEvent: &events.Event{
				Type:    events.EventTypeResources,
				Action:  events.EventActionResourcesUpdated,
				Context: context.Background(),
				Source:  []*unstructured.Unstructured{},
			},
			mockExecution: func(t *testing.T, ctrEvent *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				ctrEvent.Context = setApplyCorrelationIDContext(ctrEvent.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Do(
					func(id, path string, unstructured []*unstructured.Unstructured) {
						testutil.AssertNil(t, unstructured)
						testWg.Done()
					})
			},
		},
		"test_action_resource_deleted": {
			chanEvent: &events.Event{
				Type:    events.EventTypeResources,
				Action:  events.EventActionResourcesDeleted,
				Context: context.Background(),
				Source:  []*unstructured.Unstructured{},
			},
			mockExecution: func(t *testing.T, ctrEvent *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				ctrEvent.Context = setApplyCorrelationIDContext(ctrEvent.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Do(
					func(id, path string, unstructured []*unstructured.Unstructured) {
						testutil.AssertNil(t, unstructured)
						testWg.Done()
					})
			},
		},
		"test_action_resource_delayed_no_unstructured": {
			chanEvent: &events.Event{
				Type:    events.EventTypeResources,
				Action:  events.EventActionResourcesDeleted,
				Context: context.Background(),
				Source:  "NoUnstructured",
			},
			mockExecution: func(t *testing.T, ctrEvent *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				ctrEvent.Context = setApplyCorrelationIDContext(ctrEvent.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Do(
					func(id, path string, unstructured []*unstructured.Unstructured) {
						testutil.AssertNil(t, unstructured)
						testutil.AssertNotNil(t, testCtrOrchestrator.(*updateOrchestratorFeature).currentStateTimer)
						testWg.Done()
					})
			},
		},
		"test_action_resource_delayed_no_selfupdate": {
			chanEvent: &events.Event{
				Type:    events.EventTypeResources,
				Action:  events.EventActionResourcesDeleted,
				Context: context.Background(),
				Source: &unstructured.Unstructured{
					Object: map[string]interface{}{"kind": "test"},
				},
			},
			mockExecution: func(t *testing.T, ctrEvent *events.Event, testWg *sync.WaitGroup) {
				testWg.Add(1)
				ctrEvent.Context = setApplyCorrelationIDContext(ctrEvent.Context, testCorrelationID)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Do(
					func(id, path string, unstructured []*unstructured.Unstructured) {
						testutil.AssertNil(t, unstructured)
						testutil.AssertNotNil(t, testCtrOrchestrator.(*updateOrchestratorFeature).currentStateTimer)
						testWg.Done()
					})
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			testWg := &sync.WaitGroup{}
			testCase.mockExecution(t, testCase.chanEvent, testWg)
			eventChan <- testCase.chanEvent
			testutil.AssertWithTimeout(t, testWg, testEventsTimeout)
		})
	}
}

func setUpMocks(t *testing.T) *gomock.Controller {
	controller := gomock.NewController(t)

	setupEventsManagerMock(controller)
	setupThingMock(controller)
	setupUpdateManagerMock(controller)

	return controller
}

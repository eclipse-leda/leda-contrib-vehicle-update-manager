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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestUpdateOrchestratorCreateFeature(t *testing.T) {
	controller := gomock.NewController(t)

	setupEventsManagerMock(controller)
	setupThingMock(controller)
	setupUpdateManagerMock(controller)

	testUpdOrchestrator := newUpdateOrchestratorFeature(mockThing, mockEventsManager, mockUpdateManager)

	defer func() {
		testUpdOrchestrator.dispose()
		controller.Finish()
	}()

	dittoFeature := testUpdOrchestrator.(*updateOrchestratorFeature).createFeature()
	testutil.AssertEqual(t, UpdateOrchestratorFeatureID, dittoFeature.GetID())
	testutil.AssertNil(t, dittoFeature.GetDefinition())
	expectedProps := map[string]interface{}{
		updateOrchestratorFeaturePropertyStatus: testUpdOrchestrator.(*updateOrchestratorFeature).status,
	}
	testutil.AssertEqual(t, expectedProps, dittoFeature.GetProperties())
}

func TestUpdateOrchestratorRegister(t *testing.T) {
	controller := gomock.NewController(t)

	setupEventsManagerMock(controller)
	setupThingMock(controller)
	setupUpdateManagerMock(controller)

	testUpdOrchestrator := newUpdateOrchestratorFeature(mockThing, mockEventsManager, mockUpdateManager)

	defer func() {
		testUpdOrchestrator.dispose()
		controller.Finish()
	}()

	// init mocks
	eventChan := make(chan *events.Event)
	errorChan := make(chan error)
	mockUpdateManager.EXPECT().Get(gomock.Any()).Return(nil)
	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
	mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(1).Do(func(id string, dittoFeature model.Feature) {
		testutil.AssertEqual(t, UpdateOrchestratorFeatureID, id)
		testutil.AssertEqual(t, UpdateOrchestratorFeatureID, dittoFeature.GetID())
		testutil.AssertNil(t, dittoFeature.GetDefinition())
		expectedProps := map[string]interface{}{
			updateOrchestratorFeaturePropertyStatus: testUpdOrchestrator.(*updateOrchestratorFeature).status,
		}
		testutil.AssertEqual(t, expectedProps, dittoFeature.GetProperties())
	}).Return(nil)
	mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any())

	// test behavior
	testUpdOrchestrator.register(context.Background())
}

func TestUpdateOrchestratorOperationsHandlerNegative(t *testing.T) {
	controller := gomock.NewController(t)

	setupEventsManagerMock(controller)
	setupThingMock(controller)
	setupUpdateManagerMock(controller)

	testUpdOrchestrator := newUpdateOrchestratorFeature(mockThing, mockEventsManager, mockUpdateManager)

	defer func() {
		testUpdOrchestrator.dispose()
		controller.Finish()
	}()

	tests := map[string]struct {
		command string
		args    interface{}
	}{
		"test_unsupported_command_error": {
			command: "someRandomCommand",
		},
		"test_args_marshal_error": {
			command: updateOrchestratorFeatureOperationApply,
			args:    make(chan bool),
		},
		"test_args_no_correlation_id": {
			command: updateOrchestratorFeatureOperationApply,
			args:    map[string]interface{}{"something": 20},
		},
		"test_args_empty_correlation_id": {
			command: updateOrchestratorFeatureOperationApply,
			args:    map[string]interface{}{"correlationId": ""},
		},
		"test_args_no_payload": {
			command: updateOrchestratorFeatureOperationApply,
			args:    map[string]interface{}{"correlationId": "test-correlation-id"},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			res, err := testUpdOrchestrator.(*updateOrchestratorFeature).featureOperationsHandler(testCase.command, testCase.args)
			testutil.AssertNil(t, res)
			testutil.AssertNotNil(t, err)
		})
	}
}

func TestUpdateOrchestratorApplyCalled(t *testing.T) {
	controller := gomock.NewController(t)

	setupEventsManagerMock(controller)
	setupThingMock(controller)
	setupUpdateManagerMock(controller)

	testUpdOrchestrator := newUpdateOrchestratorFeature(mockThing, mockEventsManager, mockUpdateManager)

	defer func() {
		testUpdOrchestrator.dispose()
		controller.Finish()
	}()

	applyConfig := make(map[string]interface{})
	applyConfig["correlationId"] = "test-correlation-id"
	applyConfig["payload"] = "test-payload"

	testWg := &sync.WaitGroup{}
	testWg.Add(1)
	mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, mf []*unstructured.Unstructured) {
		testWg.Done()
	}).Times(1)

	testUpdOrchestrator.(*updateOrchestratorFeature).featureOperationsHandler(updateOrchestratorFeatureOperationApply, applyConfig)
	testutil.AssertWithTimeout(t, testWg, 5*time.Second)
}

func TestUpdateOrchestratorOperationsHandlerProcessApply(t *testing.T) {
	controller := gomock.NewController(t)

	setupEventsManagerMock(controller)
	setupThingMock(controller)
	setupUpdateManagerMock(controller)

	testUpdOrchestrator := newUpdateOrchestratorFeature(mockThing, mockEventsManager, mockUpdateManager)
	testManifest := getTestManifest()

	defer controller.Finish()

	type mockExecProcessApply func(t *testing.T, mf []*unstructured.Unstructured)
	tests := map[string]struct {
		mockExec mockExecProcessApply
	}{
		"test_process_apply_no_err": {
			mockExec: func(t *testing.T, mf []*unstructured.Unstructured) {
				eventChan := make(chan *events.Event)
				errorChan := make(chan error)
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)

				mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(1)

				mockUpdateManager.EXPECT().Get(gomock.Any())
				mockUpdateManager.EXPECT().Apply(gomock.Any(), testManifest).Times(1)

				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Times(1)
			},
		},
		"test_process_apply_err": {
			mockExec: func(t *testing.T, mf []*unstructured.Unstructured) {
				eventChan := make(chan *events.Event)
				errorChan := make(chan error)
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
				mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(1)

				mockUpdateManager.EXPECT().Get(gomock.Any())
				mockUpdateManager.EXPECT().Apply(gomock.Any(), testManifest).Return(fmt.Errorf("error apply")).Times(1)

				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any()).Times(1)
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testCase.mockExec(t, testManifest)
			ctx := context.Background()
			testUpdOrchestrator.register(ctx)
			defer testUpdOrchestrator.dispose()
			testUpdOrchestrator.(*updateOrchestratorFeature).processApply(ctx, testManifest)
		})
	}
}

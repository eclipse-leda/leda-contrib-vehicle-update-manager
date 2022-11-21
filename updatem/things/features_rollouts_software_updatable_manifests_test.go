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
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-kanto/container-management/rollouts/api/features"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCreateSUPMfFeature(t *testing.T) {
	var (
		testSuMf        managedFeature
		testSUMfFeature model.Feature
	)
	controller := gomock.NewController(t)

	setupThingMock(controller)
	setupUpdateManagerMock(controller)
	setupEventsManagerMock(controller)
	testSuMf = newSoftwareUpdatableManifests(mockThing, mockEventsManager, mockUpdateManager)

	defer func() {
		controller.Finish()
		testSuMf.dispose()
	}()

	testSUMfFeature = testSuMf.(*softwareUpdatableManifests).createFeature()
	testutil.AssertEqual(t, SoftwareUpdatableManifestsFeatureID, testSUMfFeature.GetID())
	testutil.AssertEqual(t, 1, len(testSUMfFeature.GetDefinition()))
	testutil.AssertEqual(t, client.NewDefinitionID(softwareUpdatableDefinitionNamespace, softwareUpdatableDefinitionName, softwareUpdatableDefinitionVersion).String(), testSUMfFeature.GetDefinition()[0].String())
	testutil.AssertEqual(t, 1, len(testSUMfFeature.GetProperties()))
	testutil.AssertNotNil(t, testSUMfFeature.GetProperties()[softwareUpdatablePropertyNameStatus])
	testutil.AssertEqual(t, reflect.TypeOf(&features.SoftwareUpdatableStatus{}), reflect.TypeOf(testSUMfFeature.GetProperties()[softwareUpdatablePropertyNameStatus]))
	testSUPStatus := testSUMfFeature.GetProperties()[softwareUpdatablePropertyNameStatus].(*features.SoftwareUpdatableStatus)
	testutil.AssertEqual(t, testSUPStatus.SoftwareModuleType, updateSoftwareUpdatableManifestsAgentType)
}

func TestSUMfFeatureOperationsHandler(t *testing.T) {
	const (
		testWaitTimeout = 5 * time.Second
		testMfJSON      = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: test-oci
  name: oci-app
spec:
  containers:
  - image: docker.io/library/hello-world:latest
    name: hello
`
		testInvalidJSON = `	test-invalid-json`
	)
	var (
		testSuMf        managedFeature
		testSUMfFeature model.Feature
		// test hashes
		testMfJSONHashMd5 = md5.Sum([]byte(testMfJSON))
		testMfJSONHash    = hex.EncodeToString(testMfJSONHashMd5[:])

		testInvalidJSONHashMd5 = md5.Sum([]byte(testInvalidJSON))
		testInvalidJSONHash    = hex.EncodeToString(testInvalidJSONHashMd5[:])
	)

	httpMockedCalls := map[string]func(http.ResponseWriter, *http.Request){
		testHTTPServerImageURLPathValid: func(writer http.ResponseWriter, request *http.Request) {
			// hash value of the json is testImageJSONHash
			_, _ = writer.Write([]byte(testMfJSON))

		},
		testHTTPServerImageURLPathInvalid: func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte(testInvalidJSON))
		},
	}
	setupDummyHTTPServerForTests(true, httpMockedCalls)
	defer mockHTTPServer.Close()

	testSWModule := &datatypes.SoftwareModuleID{
		Name:    testSoftwareName,
		Version: testSoftwareVersion,
	}
	assertStatesEqual := func(t *testing.T, operationStatus *datatypes.OperationStatus, expectedStatus datatypes.Status) {
		testutil.AssertEqual(t, testSWModule, operationStatus.SoftwareModule)
		testutil.AssertEqual(t, testCorrelationID, operationStatus.CorrelationID)
		testutil.AssertEqual(t, expectedStatus, operationStatus.Status)
	}

	assertApplyCtx := func(t *testing.T, correlationId, swModuleName, swModuleVersion string, ctx context.Context) {
		ctxStatus := ctx.Value(contextKeyOperationStatus)
		testutil.AssertNotNil(t, ctxStatus)
		ctxStatusInternal, ok := ctxStatus.(*suOperationContextValue)
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, correlationId, ctxStatusInternal.correlationID)
		testutil.AssertEqual(t, swModuleName, ctxStatusInternal.softwareModuleName)
		testutil.AssertEqual(t, swModuleVersion, ctxStatusInternal.softwareModuleVersion)
	}

	type mockExec func(t *testing.T) (*sync.WaitGroup, error)

	tests := map[string]struct {
		operation     string
		opts          interface{}
		mockExecution mockExec
	}{
		"test_su_feature_operations_handler_install": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: testSWModule,
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: testMfJSONHash,
								},
							},
						},
					},
				},
			},
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				eventChan := make(chan *events.Event)
				errorChan := make(chan error)
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())

				mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Do(
					func(ctx context.Context, mf []*unstructured.Unstructured) {
						assertApplyCtx(t, testCorrelationID, testSWModule.Name, testSWModule.Version, ctx)
						eventChan <- &events.Event{Type: orchestration.EventTypeOrchestration, Action: orchestration.EventActionOrchestrationStarted, Context: ctx}
						eventChan <- &events.Event{Type: orchestration.EventTypeOrchestration, Action: orchestration.EventActionOrchestrationRunning, Context: ctx}
						eventChan <- &events.Event{Type: orchestration.EventTypeOrchestration, Action: orchestration.EventActionOrchestrationFinished, Context: ctx}
					}).Times(1)

				wg := &sync.WaitGroup{}
				wg.Add(3) // Installing, Installed, Finished Success
				gomock.InOrder(
					// Started
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Started)
					}).Times(1).Return(nil),
					// Downloading
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Downloading)
					}).Times(1).Return(nil),
					// Downloaded
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Downloaded)
					}).Times(1).Return(nil),
					// Installing
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Installing)
						wg.Done()
					}).Times(1).Return(nil),
					// Installed
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Installed)
						wg.Done()
					}).Times(1).Return(nil),
					// Finished Success
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedSuccess)
						wg.Done()
					}).Times(1).Return(nil),
				)
				return wg, nil
			},
		},
		"test_su_feature_operations_handler_install_invalid_hash": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: testSWModule,
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: "invalid-hash",
								},
							},
						},
					},
				},
			},
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())
				mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Times(0)
				wg := &sync.WaitGroup{}
				wg.Add(1)
				gomock.InOrder(
					// Started
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Started)
					}).Times(1).Return(nil),
					// Downloading
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Downloading)
					}).Times(1).Return(nil),
					// Failed
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastFailedOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedRejected)
					}).Times(1).Return(nil),
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedRejected)
						wg.Done()
					}).Times(1).Return(nil),
				)
				return wg, nil
			},
		},
		"test_su_feature_operations_handler_install_invalid_artifact": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: testSWModule,
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathInvalid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: testInvalidJSONHash,
								},
							},
						},
					},
				},
			},
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())
				mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Times(0)
				wg := &sync.WaitGroup{}
				wg.Add(1)
				gomock.InOrder(
					// Started
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Started)
					}).Times(1).Return(nil),
					// Downloading
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Downloading)
					}).Times(1).Return(nil),
					// Failed
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastFailedOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedError)
					}).Times(1).Return(nil),
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedError)
						wg.Done()
					}).Times(1).Return(nil),
				)
				return wg, nil
			},
		},
		"test_su_feature_operations_handler_install_panic": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: testSWModule,
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: testMfJSONHash,
								},
							},
						},
					},
				},
			},
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())
				mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Times(0)
				wg := &sync.WaitGroup{}
				wg.Add(1)
				gomock.InOrder(
					// Started
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Started)
					}).Times(1).Return(nil),
					// Downloading
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Downloading)
						panic(log.NewError("test error"))
					}).Times(1).Return(nil),
					// Finished Error
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastFailedOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedError)
					}).Times(1).Return(nil),
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedError)
						wg.Done()
					}).Times(1).Return(nil),
				)
				return wg, nil
			},
		},
		"test_su_feature_operations_handler_install_no_artifacts": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: testSWModule,
						Artifacts:      []*datatypes.SoftwareArtifactAction{},
					},
				}},
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())
				mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Times(0)
				mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), gomock.Any(), gomock.Any()).Times(0).Return(nil)
				return nil, client.NewMessagesParameterInvalidError("the number of SoftwareArtifacts referenced for SoftwareModule [Name.version] = [%s.%s] must be exactly 1", testSWModule.Name, testSWModule.Version)
			},
		},
		"test_su_feature_operations_handler_install_no_modules": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				SoftwareModules: []*datatypes.SoftwareModuleAction{},
			},
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())
				mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Times(0)
				mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), gomock.Any(), gomock.Any()).Times(0).Return(nil)
				return nil, client.NewMessagesParameterInvalidError("the number of SoftwareModules must be exactly 1")
			},
		},
		"test_su_feature_operations_handler_install_error_installing": {
			operation: softwareUpdatableOperationInstall,
			opts: datatypes.UpdateAction{
				CorrelationID: testCorrelationID,
				SoftwareModules: []*datatypes.SoftwareModuleAction{
					{
						SoftwareModule: testSWModule,
						Artifacts: []*datatypes.SoftwareArtifactAction{
							{
								Download: map[datatypes.Protocol]*datatypes.Links{
									datatypes.HTTP: {
										URL: mockHTTPServer.URL + testHTTPServerImageURLPathValid,
									},
								},
								Checksums: map[datatypes.Hash]string{
									datatypes.MD5: testMfJSONHash,
								},
							},
						},
					},
				},
			},
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				eventChan := make(chan *events.Event)
				errorChan := make(chan error)
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())

				mockUpdateManager.EXPECT().Apply(gomock.Any(), gomock.Any()).Do(
					func(ctx context.Context, mf []*unstructured.Unstructured) {
						assertApplyCtx(t, testCorrelationID, testSWModule.Name, testSWModule.Version, ctx)
						eventChan <- &events.Event{Type: orchestration.EventTypeOrchestration, Action: orchestration.EventActionOrchestrationStarted, Context: ctx}
						eventChan <- &events.Event{Type: orchestration.EventTypeOrchestration, Action: orchestration.EventActionOrchestrationRunning, Context: ctx}
						eventChan <- &events.Event{Type: orchestration.EventTypeOrchestration, Action: orchestration.EventActionOrchestrationFinished, Context: ctx, Error: log.NewError("test error")}
					}).Times(1)

				wg := &sync.WaitGroup{}
				wg.Add(3) // Installing, Finished Error - last operation, last failed operation
				gomock.InOrder(
					// Started
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Started)
					}).Times(1).Return(nil),
					// Downloading
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Downloading)
					}).Times(1).Return(nil),
					// Downloaded
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Downloaded)
					}).Times(1).Return(nil),
					// Installing
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.Installing)
						wg.Done()
					}).Times(1).Return(nil),
					// Finished Error
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastFailedOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedError)
						wg.Done()
					}).Times(1).Return(log.NewError("test error")),
					mockThing.EXPECT().SetFeatureProperty(testSUMfFeature.GetID(), softwareUpdatablePropertyLastOperation, gomock.Any()).Do(func(fId, propertyPath string, status *datatypes.OperationStatus) {
						assertStatesEqual(t, status, datatypes.FinishedError)
						wg.Done()
					}).Times(1).Return(log.NewError("test error")),
				)
				return wg, nil
			},
		},
		"test_su_feature_operations_handler_default": {
			operation: "unsupportedOperation",
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())
				return nil, client.NewMessagesSubjectNotFound(log.NewErrorf("unsupported operation called [operationId = %s]", "unsupportedOperation").Error())
			},
		},
		"test_su_feature_operations_handler_invalid_args": {
			operation: softwareUpdatableOperationInstall,
			opts:      make(chan int),
			mockExecution: func(t *testing.T) (*sync.WaitGroup, error) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any())
				return nil, client.NewMessagesParameterInvalidError("json: unsupported type: chan int")
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			controller := gomock.NewController(t)
			setupThingMock(controller)
			setupUpdateManagerMock(controller)
			setupEventsManagerMock(controller)

			testSuMf = newSoftwareUpdatableManifests(mockThing, mockEventsManager, mockUpdateManager)
			testSUMfFeature = testSuMf.(*softwareUpdatableManifests).createFeature()

			defer func() {
				controller.Finish()
				testSuMf.dispose()
			}()

			testWg, expectedRunErr := testCase.mockExecution(t)
			regErr := testSuMf.register(context.Background())
			testutil.AssertNil(t, regErr)

			result, resultErr := testSuMf.(*softwareUpdatableManifests).operationsHandler(testCase.operation, testCase.opts)
			if testWg != nil {
				testutil.AssertWithTimeout(t, testWg, testWaitTimeout)
			}
			testutil.AssertNil(t, result)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

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
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"

	mocksevents "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/events"
	mocksorchmgr "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/orchestration"
	mocksupdorchmgr "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/updateorchestrator"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/selfupdate"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"

	"github.com/golang/mock/gomock"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
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

const selfUpdateMultipleBundlesManifest = `
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
  name: self-update-bundle-example-3
spec:
  bundleName: swdv-arm64-build42
  bundleVersion: v1beta3
  bundleDownloadUrl: https://example.com/repository/base/
  bundleTarget: base
---
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
  name: self-update-bundle-example-3
spec:
  bundleName: swdv-arm64-build42
  bundleVersion: v1beta3
  bundleDownloadUrl: https://example.com/repository/base/
  bundleTarget: base
---
`

const manifest = `
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
  name: self-update-bundle-example-3
spec:
  bundleName: swdv-arm64-build42
  bundleVersion: v1beta3
  bundleDownloadUrl: https://example.com/repository/base/
  bundleTarget: base
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
  matchLabels:
    app: nginx
  template:
    metadata:
      labels:
      app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
          - containerPort: 80
`
const k8sManifest = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
  matchLabels:
    app: nginx
  template:
    metadata:
      labels:
      app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
          - containerPort: 80
`

func TestApply(t *testing.T) {
	type mockFunc func(*mocksevents.MockUpdateEventsManager, *mocksorchmgr.MockUpdateManager, *mocksorchmgr.MockUpdateManager, *mocksupdorchmgr.MockRebootManager, []*unstructured.Unstructured)

	tests := []struct {
		testName      string
		manifest      string
		mockExecution mockFunc
	}{
		{
			"apply_self_update_manifest",
			selfUpdateManifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				setupEventsManager(t, mockEventsMgr, mf, nil)
				mockSelfUpdateMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&selfupdate.ApplyResult{})
			},
		},
		{
			"apply_self_update_manifest_reboot_required",
			selfUpdateManifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				setupEventsManager(t, mockEventsMgr, mf, nil)
				mockRebootMgr.EXPECT().Reboot(gomock.Any())
				mockSelfUpdateMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&selfupdate.ApplyResult{RebootRequired: true})
			},
		},
		{
			"apply_self_update_error",
			selfUpdateManifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				applyErr := fmt.Errorf("error applying self update manifest")
				setupEventsManager(t, mockEventsMgr, mf, applyErr)
				mockSelfUpdateMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&selfupdate.ApplyResult{Err: applyErr})
			},
		},
		{
			"apply_self_update_multiple_bundles",
			selfUpdateMultipleBundlesManifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				setupEventsManager(t, mockEventsMgr, mf, fmt.Errorf("more than one SelfUpdateBundle resource in the YAML manifest"))
			},
		},
		{
			"apply_k8s_manifest",
			k8sManifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				setupEventsManager(t, mockEventsMgr, mf, nil)
				mockK8sOrchestrationMgr.EXPECT().Apply(gomock.Any(), gomock.Any())
			},
		},
		{
			"apply_k8s_manifest_error",
			k8sManifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				applyErr := fmt.Errorf("error applying k8s manifest")
				setupEventsManager(t, mockEventsMgr, mf, applyErr)
				mockK8sOrchestrationMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(applyErr)
			},
		},
		{
			"apply_k8s_and_self_update_manifest",
			manifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				setupEventsManager(t, mockEventsMgr, mf, nil)
				mockSelfUpdateMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&selfupdate.ApplyResult{})
				mockK8sOrchestrationMgr.EXPECT().Apply(gomock.Any(), gomock.Any())
			},
		},
		{
			"apply_k8s_and_self_update_manifest_reboot_required",
			manifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				setupEventsManager(t, mockEventsMgr, mf, nil)
				mockRebootMgr.EXPECT().Reboot(gomock.Any())
				mockSelfUpdateMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&selfupdate.ApplyResult{RebootRequired: true})
				mockK8sOrchestrationMgr.EXPECT().Apply(gomock.Any(), gomock.Any())
			},
		},
		{
			"apply_k8s_and_self_update_manifest_error",
			manifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				applyErr := fmt.Errorf("error applying self update manifest")
				setupEventsManager(t, mockEventsMgr, mf, applyErr)
				mockSelfUpdateMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&selfupdate.ApplyResult{Err: applyErr})
			},
		},
		{
			"apply_k8s_and_self_update_error_reboot_required",
			manifest,
			func(mockEventsMgr *mocksevents.MockUpdateEventsManager, mockSelfUpdateMgr *mocksorchmgr.MockUpdateManager, mockK8sOrchestrationMgr *mocksorchmgr.MockUpdateManager, mockRebootMgr *mocksupdorchmgr.MockRebootManager, mf []*unstructured.Unstructured) {
				applyErr := fmt.Errorf("error applying k8s manifest")
				setupEventsManager(t, mockEventsMgr, mf, applyErr)
				mockRebootMgr.EXPECT().Reboot(gomock.Any())
				mockSelfUpdateMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(&selfupdate.ApplyResult{RebootRequired: true})
				mockK8sOrchestrationMgr.EXPECT().Apply(gomock.Any(), gomock.Any()).Return(applyErr)
			},
		},
	}

	controller := gomock.NewController(t)
	for _, testValues := range tests {
		t.Run(testValues.testName, func(t *testing.T) {
			mockEventsMgr := mocksevents.NewMockUpdateEventsManager(controller)
			mockRebootMgr := mocksupdorchmgr.NewMockRebootManager(controller)
			mockSelfUpdateMgr := mocksorchmgr.NewMockUpdateManager(controller)
			mockK8sOrchestrationMgr := mocksorchmgr.NewMockUpdateManager(controller)

			_, mf, _ := parseMultiYAML([]byte(testValues.manifest))
			testValues.mockExecution(mockEventsMgr, mockSelfUpdateMgr, mockK8sOrchestrationMgr, mockRebootMgr, mf)
			orchMgr := createTestUpdateOrchestrator(mockEventsMgr, mockSelfUpdateMgr, mockK8sOrchestrationMgr, mockRebootMgr)
			orchMgr.Apply(context.Background(), mf)
		})
	}
}

func TestGet(t *testing.T) {
	controller := gomock.NewController(t)

	mockRebootMgr := mocksupdorchmgr.NewMockRebootManager(controller)
	mockSelfUpdateMgr := mocksorchmgr.NewMockUpdateManager(controller)
	mockK8sOrchestrationMgr := mocksorchmgr.NewMockUpdateManager(controller)

	_, k8sMf, _ := parseMultiYAML([]byte(k8sManifest))
	_, selfUpdateMf, _ := parseMultiYAML([]byte(selfUpdateManifest))

	mockSelfUpdateMgr.EXPECT().Get(gomock.Any()).Return(selfUpdateMf)
	mockK8sOrchestrationMgr.EXPECT().Get(gomock.Any()).Return(k8sMf)

	orchMgr := createTestUpdateOrchestrator(nil, mockSelfUpdateMgr, mockK8sOrchestrationMgr, mockRebootMgr)
	result := orchMgr.Get(context.Background())

	if len(result) != 2 {
		t.Fail()
	}
	if result[0].GetKind() != "Deployment" {
		t.Fail()
	}
	if result[1].GetKind() != "SelfUpdateBundle" {
		t.Fail()
	}
	testutil.AssertNil(t, orchMgr.Dispose(context.Background()))
}

func TestShouldReturnErrorOnReboot(t *testing.T) {
	updateOrchestrator := &rebootManager{}
	err := updateOrchestrator.Reboot(2000)

	testutil.AssertError(t, log.NewError("cannot reboot after successful update operation. cannot send signal to /proc/sys/kernel/sysrq"), err)
}

func parseMultiYAML(multiYamlData []byte) ([][]byte, []*unstructured.Unstructured, error) {
	mf := []*unstructured.Unstructured{}
	singleYamlDoc, yamlReadErr := readResources(multiYamlData)
	if yamlReadErr != nil {
		return nil, nil, yamlReadErr
	}
	if len(singleYamlDoc) == 0 {
		return nil, nil, log.NewError("no k8s content provided for processing")
	}

	for _, pod := range singleYamlDoc {
		bs, yamlErr := yaml.YAMLToJSON(pod)
		if yamlErr != nil {
			log.ErrorErr(yamlErr, "error reading yaml content")
			return nil, nil, yamlErr
		}
		podSpec := &unstructured.Unstructured{}
		podSpec.UnmarshalJSON(bs)
		mf = append(mf, podSpec)
	}
	log.Debug("downloaded update manifest to install from Things:\n%+v", mf)
	return singleYamlDoc, mf, nil
}

func readResources(multiYamlData []byte) ([][]byte, error) {
	var documentList [][]byte
	yamlDecoder := yamlv3.NewDecoder(bytes.NewReader(multiYamlData))
	for {
		var singleDoc interface{}
		err := yamlDecoder.Decode(&singleDoc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, log.NewErrorf("error splitting multi yaml content:\n%+v", err)
		}

		if singleDoc != nil {
			document, marshalErr := yamlv3.Marshal(singleDoc)
			if marshalErr != nil {
				return nil, log.NewErrorf("error marshalling single read yaml document:\n%+v", marshalErr)
			}
			documentList = append(documentList, document)
		}
	}
	return documentList, nil
}

func assertError(t *testing.T, expected error, actual error) {
	if expected == nil {
		if actual != nil {
			t.Errorf("expected nil , got %v", actual)
			t.Fail()
		}
	} else {
		if actual == nil {
			t.Errorf("expected %v , got nil", expected)
			t.Fail()
		}
	}
}

func createTestUpdateOrchestrator(mockEventsMgr events.UpdateEventsManager, mockSelfUpdateMgr, mockK8sOrchestrationMgr orchestration.UpdateManager,
	mockRebootMgr RebootManager) orchestration.UpdateManager {
	return &updateOrchestrator{
		applyLock:               sync.Mutex{},
		rebootManager:           mockRebootMgr,
		eventsManager:           mockEventsMgr,
		selfUpdateManager:       mockSelfUpdateMgr,
		k8sOrchestrationManager: mockK8sOrchestrationMgr,
	}
}

func setupEventsManager(t *testing.T, mockEventsMgr *mocksevents.MockUpdateEventsManager, mf []*unstructured.Unstructured, expectedErr error) {
	assertCtx := func(ctx context.Context, mf []*unstructured.Unstructured) {
		testutil.AssertEqual(t, mf, orchestration.GetUpdateMgrApplyContext(ctx))
	}
	assertEvent := func(event *events.Event, expAction events.EventAction, expErr error) {
		testutil.AssertEqual(t, orchestration.EventTypeOrchestration, event.Type)
		testutil.AssertEqual(t, expAction, event.Action)
		assertError(t, expErr, event.Error)
	}
	gomock.InOrder(
		mockEventsMgr.EXPECT().Publish(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, event *events.Event) {
			assertCtx(ctx, mf)
			assertCtx(event.Context, mf)
			assertEvent(event, orchestration.EventActionOrchestrationStarted, nil)
		},
		).Return(nil),
		mockEventsMgr.EXPECT().Publish(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, event *events.Event) {
			assertCtx(ctx, mf)
			assertCtx(event.Context, mf)
			assertEvent(event, orchestration.EventActionOrchestrationFinished, expectedErr)
		},
		).Return(nil),
	)
}

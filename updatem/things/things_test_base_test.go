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
	"net/http"
	"net/http/httptest"

	mocksevents "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/events"
	mocksorchmgr "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/orchestration"
	mocksthings "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/things"
	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testHTTPServerImageURLPathValid   = "/image/valid"
	testHTTPServerImageURLPathInvalid = "/image/invalid"
	testSoftwareName                  = "image:latest"
	testSoftwareVersion               = "1.0"
	testCorrelationID                 = "1000"
)

var (
	testThingsFeaturesDefaultSet = []string{SoftwareUpdatableManifestsFeatureID}
	mockEventsManager            *mocksevents.MockUpdateEventsManager
	mockThing                    *mocksthings.MockThing
	mockUpdateManager            *mocksorchmgr.MockUpdateManager
	testThingsMgr                *updateThingsMgr
	mockHTTPServer               *httptest.Server
)

func setupEventsManagerMock(controller *gomock.Controller) {
	mockEventsManager = mocksevents.NewMockUpdateEventsManager(controller)
}

func setupThingMock(controller *gomock.Controller) {
	mockThing = mocksthings.NewMockThing(controller)
}

func setupUpdateManagerMock(controller *gomock.Controller) {
	mockUpdateManager = mocksorchmgr.NewMockUpdateManager(controller)
}

const (
	testThingsStoragePath = "../pkg/testutil/metapath/valid/things"
)

func setupThingsUpdateManager(controller *gomock.Controller) {
	testThingsMgr = newThingsUpdateManager(
		mockUpdateManager,
		mockEventsManager,
		"",
		0,
		0,
		"",
		"",
		testThingsStoragePath,
		testThingsFeaturesDefaultSet,
		0,
		0,
		0,
		0,
	)
}

/*
	Basically the expected outgoing http request could be asserted by either

creating a wrapper http client and mocking it OR by mocking an http server.
The second approach looks cleaner at the moment,
as it does not require changes in the source code.
*/
func setupDummyHTTPServerForTests(plain bool, mockedCalls map[string]func(http.ResponseWriter, *http.Request)) {
	handler := http.NewServeMux()
	for path, handlerFunc := range mockedCalls {
		handler.HandleFunc(path, handlerFunc)
	}
	if plain {
		mockHTTPServer = httptest.NewServer(handler)
		return
	}
	mockHTTPServer = httptest.NewTLSServer(handler)
}

func getTestManifest() []*unstructured.Unstructured {
	const (
		testPodName = "test-app"
		testImage   = "test.test/image:latest"
		testName    = "test-name"
	)

	unstructContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: testPodName},
		Spec: v1.PodSpec{
			Containers: []v1.Container{v1.Container{Name: testName, Image: testImage}},
		},
		Status: v1.PodStatus{},
	})
	unstructPod := &unstructured.Unstructured{}
	unstructPod.SetUnstructuredContent(unstructContent)
	return []*unstructured.Unstructured{unstructPod}
}

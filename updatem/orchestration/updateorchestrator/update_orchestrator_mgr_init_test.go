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
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/k8s"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/selfupdate"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	eventsmock "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/events"

	"github.com/eclipse-kanto/container-management/containerm/registry"
	registryservices "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/registry"

	"github.com/golang/mock/gomock"
)

func TestRegistryInit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	eventsMgrMock := eventsmock.NewMockUpdateEventsManager(mockCtrl)
	orchRegistration := &registry.Registration{
		ID:       UpdateOrchestratorServiceID,
		Type:     registryservices.UpdateManagerService,
		InitFunc: registryInit,
	}
	testCases := map[string]struct {
		testServiceInfoSet func() *registry.Set
		config             interface{}
	}{
		"test_bad_config": {
			testServiceInfoSet: func() *registry.Set {
				return nil
			},
			config: "some-test-cfg",
		},
		"test_missing_k8s_config": {
			testServiceInfoSet: func() *registry.Set {
				return nil
			},
			config: map[string]interface{}{
				"self_update": []selfupdate.MgrOpt{},
			},
		},
		"test_missing_selfupdate_config": {
			testServiceInfoSet: func() *registry.Set {
				return nil
			},
			config: map[string]interface{}{
				"k8s": []k8s.MgrOpt{
					k8s.WithKubeConfig("../../pkg/testutil/testdata/k8s/k3s.yaml"),
				},
			},
		},
		"test_no_events_mgr_service": {
			testServiceInfoSet: func() *registry.Set {
				return registry.NewServiceInfoSet()
			},
			config: map[string]interface{}{
				"k8s": []k8s.MgrOpt{
					k8s.WithKubeConfig("../../pkg/testutil/testdata/k8s/k3s.yaml"),
				},
				"self_update": []selfupdate.MgrOpt{},
			},
		},
		"test_corrupted_config": {
			testServiceInfoSet: func() *registry.Set {
				serviceInfoSet := registry.NewServiceInfoSet()

				evenetsMgrRegistration := &registry.Registration{
					ID:   events.EventsManagerServiceLocalID,
					Type: registry.EventsManagerService,
					InitFunc: func(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
						return eventsMgrMock, nil
					},
				}
				registry.Register(evenetsMgrRegistration)
				serviceInfo := evenetsMgrRegistration.Init(registry.NewContext(context.Background(), nil, evenetsMgrRegistration, serviceInfoSet))

				serviceInfoSet.Add(serviceInfo)
				return serviceInfoSet
			},
			config: map[string]interface{}{
				"k8s": []k8s.MgrOpt{
					k8s.WithKubeConfig("../../pkg/testutil/testdata/k8s/k3s_corrupted.yaml"),
				},
				"self_update": []selfupdate.MgrOpt{},
			},
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			serviceSet := testCase.testServiceInfoSet()
			orchServiceInfo := orchRegistration.Init(registry.NewContext(context.Background(), testCase.config, orchRegistration, serviceSet))
			instance, err := orchServiceInfo.Instance()
			testutil.AssertNil(t, instance)
			testutil.AssertNotNil(t, err)
		})
	}
}

func TestCreateNewUpdateOrchestratorShouldWorkCorrectlyWithEmptyDependencies(t *testing.T) {
	updateOrchestrator := newUpdateOrchestrator(nil, nil, nil, nil, nil, nil)
	testutil.AssertNotNil(t, updateOrchestrator)
	testutil.AssertNil(t, updateOrchestrator.cfg)
	testutil.AssertNil(t, updateOrchestrator.eventsManager)
	testutil.AssertNil(t, updateOrchestrator.k8sOrchestrationManager)
	testutil.AssertNil(t, updateOrchestrator.pahoClient)
	testutil.AssertNil(t, updateOrchestrator.rebootManager)
	testutil.AssertNil(t, updateOrchestrator.selfUpdateManager)
}
func TestRegistration(t *testing.T) {
	registrationsMap := registry.RegistrationsMap()
	testutil.AssertNotNil(t, registrationsMap)
	registrations := registrationsMap[registryservices.UpdateManagerService]
	testutil.AssertEqual(t, 1, len(registrations))
	testutil.AssertEqual(t, UpdateOrchestratorServiceID, registrations[0].ID)
}

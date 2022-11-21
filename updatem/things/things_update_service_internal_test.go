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
	"testing"

	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/golang/mock/gomock"
)

func TestProcessUpdateThingDefault(t *testing.T) {
	tests := map[string]struct {
		enabledFeatures []string
		mockExec        func(t *testing.T)
	}{
		"test_default_config": {
			enabledFeatures: testThingsFeaturesDefaultSet,
			mockExec: func(t *testing.T) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(nil, nil)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any()).Times(1).Return(nil)
				mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(0).Return(nil)
			},
		},
		"test_factory_only_config": {
			enabledFeatures: []string{UpdateOrchestratorFeatureID},
			mockExec: func(t *testing.T) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(nil, nil)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any()).Times(0).Return(nil)
				mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(1).Return(nil)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any())
			},
		},
		"test_su_only_config": {
			enabledFeatures: []string{SoftwareUpdatableManifestsFeatureID},
			mockExec: func(t *testing.T) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(nil, nil)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any()).Times(1).Return(nil)
				mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(0).Return(nil)
			},
		},
		"test_all_config": {
			enabledFeatures: []string{UpdateOrchestratorFeatureID, SoftwareUpdatableManifestsFeatureID},
			mockExec: func(t *testing.T) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(2).Return(nil, nil)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any()).Times(1).Return(nil)
				mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(1).Return(nil)
				mockThing.EXPECT().SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, gomock.Any())
			},
		},
		"test_none_config": {
			enabledFeatures: nil,
			mockExec: func(t *testing.T) {
				mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(0).Return(nil, nil)
				mockThing.EXPECT().SetFeature(SoftwareUpdatableManifestsFeatureID, gomock.Any()).Times(0).Return(nil)
				mockThing.EXPECT().SetFeature(UpdateOrchestratorFeatureID, gomock.Any()).Times(0).Return(nil)
			},
		},
	}
	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)
			defer controller.Finish()
			setupEventsManagerMock(controller)
			setupThingMock(controller)
			setupThingsUpdateManager(controller)

			mockUpdateManager.EXPECT().Get(gomock.Any()).Return(nil)

			namespaceID := client.NewNamespacedID("things.update.service", "test")
			testThingsMgr.updateThingID = namespaceID.String()
			mockThing.EXPECT().GetID().Times(1).Return(namespaceID)
			testThingsMgr.enabledFeatureIds = testCase.enabledFeatures
			testCase.mockExec(t)
			testThingsMgr.processThing(mockThing)
			if testCase.enabledFeatures != nil {
				testutil.AssertEqual(t, len(testCase.enabledFeatures), len(testThingsMgr.managedFeatures))
				for _, fID := range testCase.enabledFeatures {
					testutil.AssertNotNil(t, testThingsMgr.managedFeatures[fID])
				}
			}
		})
	}
}

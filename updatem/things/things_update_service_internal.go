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

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
)

func (tMgr *updateThingsMgr) registryHandler(changedType handlers.ThingsRegistryChangedType, thing model.Thing) {
	tMgr.initMutex.Lock()
	defer tMgr.initMutex.Unlock()
	if changedType == handlers.Added {
		log.Debug("added new thing from Things service >>>>> %v", thing)
		tMgr.processThing(thing)
	}
}

func (tMgr *updateThingsMgr) thingsClientInitializedHandler(cl *client.Client, configuration *client.Configuration, err error) {
	tMgr.initMutex.Lock()
	defer tMgr.initMutex.Unlock()
	log.Debug("received things client initialized notification with client configuration: %s and Error info: %s", configuration, err)
	if err != nil {
		log.ErrorErr(err, "Error initializing things client")
		return
	}
	log.Debug("processing things client configuration")
	log.Debug("successfully initialized things manager info with {rootDeviceId:%s,rootDeviceTenantId:%s,rootDeviceAuthId:%s,rootDevicePassword:%s}", configuration.GatewayDeviceID(), configuration.DeviceTenantID(), configuration.DeviceAuthID(), configuration.DevicePassword())

	namespaceID := client.NewNamespacedIDFromString(client.NewNamespacedID(configuration.GatewayDeviceID(), configuration.DeviceName()).String())
	tMgr.updateThingID = namespaceID.String()
	rootThing := tMgr.thingsClient.Get(namespaceID)
	if rootThing == nil {
		log.Error("the root thing device with id = %s is missing in the things client's cache", tMgr.updateThingID)
	} else {
		// add features
		tMgr.processThing(rootThing)
	}
	cl.SetThingsRegistryChangedHandler(tMgr.registryHandler)
	//profile.DumpMem()
}

func (tMgr *updateThingsMgr) processThing(thing model.Thing) {
	if thing.GetID().String() == tMgr.updateThingID {
		ctx := context.Background()

		// dispose all features(their event handlers would be closed)
		tMgr.disposeFeatures()
		tMgr.managedFeatures = make(map[string]managedFeature)

		// handle UpdateOrchestrator
		if tMgr.isFeatureEnabled(UpdateOrchestratorFeatureID) {
			log.Debug("registering %s feature", UpdateOrchestratorFeatureID)
			updOrchestrator := newUpdateOrchestratorFeature(thing, tMgr.eventsMgr, tMgr.updOrchMgr)
			tMgr.managedFeatures[UpdateOrchestratorFeatureID] = updOrchestrator
		} else {
			log.Debug("%s feature is NOT enabled and will not be registered", UpdateOrchestratorFeatureID)
		}

		// handle SoftwareUpdatable:manifest
		if tMgr.isFeatureEnabled(SoftwareUpdatableManifestsFeatureID) {
			log.Debug("registering %s feature", SoftwareUpdatableManifestsFeatureID)
			suMf := newSoftwareUpdatableManifests(thing, tMgr.eventsMgr, tMgr.updOrchMgr)
			tMgr.managedFeatures[SoftwareUpdatableManifestsFeatureID] = suMf
		} else {
			log.Debug("%s feature is NOT enabled and will not be registered", SoftwareUpdatableManifestsFeatureID)
		}

		// register all added features
		for featureID, feature := range tMgr.managedFeatures {
			log.Debug("registering feature %s", featureID)
			if err := feature.register(ctx); err != nil {
				log.ErrorErr(err, "could not register %s feature", featureID)
			}
		}
	} else {
		log.Debug("the thing is not the update thing - will not process it")
	}
}

func (tMgr *updateThingsMgr) disposeFeatures() {
	for featureID, feature := range tMgr.managedFeatures {
		log.Debug("disposing feature %s", featureID)
		feature.dispose()
	}
}

func (tMgr *updateThingsMgr) isFeatureEnabled(featureID string) bool {
	if tMgr.enabledFeatureIds == nil || len(tMgr.enabledFeatureIds) == 0 {
		return false
	}
	for _, enabled := range tMgr.enabledFeatureIds {
		if enabled == featureID {
			return true
		}
	}
	return false
}

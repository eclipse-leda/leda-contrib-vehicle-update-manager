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
	"time"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"

	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	registryservices "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/registry"
)

const (
	updateThingName = "edge:update"
)

func newThingsUpdateManager(mgr orchestration.UpdateManager, eventsMgr events.UpdateEventsManager,
	brokerURL string,
	keepAlive time.Duration,
	disconnectTimeout time.Duration,
	username string,
	password string,
	storagePath string,
	enabledFeatureIds []string,
	connectTimeout time.Duration,
	acknowledgeTimeout time.Duration,
	subscribeTimeout time.Duration,
	unsubscribeTimeout time.Duration) *updateThingsMgr {
	thingsMgr := &updateThingsMgr{
		storageRoot:       storagePath,
		updOrchMgr:        mgr,
		eventsMgr:         eventsMgr,
		enabledFeatureIds: enabledFeatureIds,
		managedFeatures:   map[string]managedFeature{},
	}

	thingsClientOpts := client.NewConfiguration()
	thingsClientOpts.WithBroker(brokerURL).
		WithDisconnectTimeout(disconnectTimeout).
		WithKeepAlive(keepAlive).
		WithClientUsername(username).
		WithClientPassword(password).
		WithInitHook(thingsMgr.thingsClientInitializedHandler).
		WithDeviceName(updateThingName).
		WithConnectTimeout(connectTimeout).
		WithAcknowledgeTimeout(acknowledgeTimeout).
		WithSubscribeTimeout(subscribeTimeout).
		WithUnsubscribeTimeout(unsubscribeTimeout)

	thingsMgr.thingsClient = client.NewClient(thingsClientOpts)

	return thingsMgr
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	eventsMgr, err := registryCtx.Get(registry.EventsManagerService)
	if err != nil {
		return nil, err
	}

	mgrService, err := registryCtx.Get(registryservices.UpdateManagerService)
	if err != nil {
		return nil, err
	}

	// init options processing
	initOpts := registryCtx.Config.([]UpdateThingsManagerOpt)
	tOpts := &thingsOpts{}
	err = applyOptsThings(tOpts, initOpts...)
	if err != nil {
		return nil, err
	}
	return newThingsUpdateManager(mgrService.(orchestration.UpdateManager), eventsMgr.(events.UpdateEventsManager),
		tOpts.broker,
		tOpts.keepAlive,
		tOpts.disconnectTimeout,
		tOpts.clientUsername,
		tOpts.clientPassword,
		tOpts.storagePath,
		tOpts.featureIds,
		tOpts.connectTimeout,
		tOpts.acknowledgeTimeout,
		tOpts.subscribeTimeout,
		tOpts.unsubscribeTimeout), nil
}

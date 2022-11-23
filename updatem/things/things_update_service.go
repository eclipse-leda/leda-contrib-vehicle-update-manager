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

	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	registryservices "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/registry"
)

const (
	// ThingsUpdateManagerServiceLocalID is the ID of the things management service in the registry
	ThingsUpdateManagerServiceLocalID = "updatemanagerd.service.local.v1.service-things-update-manager"
	thingsBackupFileNameTemplate      = "sup_back_%s.json"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       ThingsUpdateManagerServiceLocalID,
		Type:     registryservices.ThingsUpdateManagerService,
		InitFunc: registryInit,
	})
}

type managedFeature interface {
	register(ctx context.Context) error
	dispose()
}

// UpdateThingsManager is the things management abstraction
type UpdateThingsManager interface {
	Connect() error
	Disconnect()
}

type updateThingsMgr struct {
	enabledFeatureIds []string
	storageRoot       string
	updOrchMgr        orchestration.UpdateManager
	eventsMgr         events.UpdateEventsManager

	thingsClient *client.Client

	updateThingID   string
	managedFeatures map[string]managedFeature
	initMutex       sync.Mutex
}

func (tMgr *updateThingsMgr) Connect() error {

	if err := tMgr.thingsClient.Connect(); err != nil {
		return err
	}

	return nil
}

func (tMgr *updateThingsMgr) Disconnect() {
	tMgr.initMutex.Lock()
	defer tMgr.initMutex.Unlock()

	tMgr.disposeFeatures()
	tMgr.thingsClient.Disconnect()
}

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

package selfupdate

import (
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/util"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

// NewSelfUpdateManager instantiates a new self update manager
func NewSelfUpdateManager(opts []MgrOpt, registryCtx *registry.ServiceRegistryContext) (orchestration.UpdateManager, error) {
	var (
		cfg = &mgrOpts{}
	)
	applyOptsMgr(cfg, opts...)

	eventsMgr, err := registryCtx.Get(registry.EventsManagerService)
	if err != nil {
		return nil, err
	}

	pahoOpts := mqtt.NewClientOptions().
		AddBroker(cfg.broker).
		SetClientID(uuid.New().String()).
		SetKeepAlive(cfg.keepAlive).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectTimeout(cfg.connectTimeout)
	pahoClient := mqtt.NewClient(pahoOpts)

	if err := util.MqttConnect(pahoClient, cfg.broker); err != nil {
		return nil, err
	}

	suMgr := &selfUpdateManager{
		cfg:        cfg,
		pahoClient: pahoClient,
		eventsMgr:  eventsMgr.(events.UpdateEventsManager),
	}

	if err := suMgr.subscribeSelfUpdateCurrentState(); err != nil {
		return nil, err
	}

	return suMgr, nil
}

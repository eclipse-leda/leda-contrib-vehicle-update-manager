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
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/k8s"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/selfupdate"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/util"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	registryservices "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/registry"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       UpdateOrchestratorServiceID,
		Type:     registryservices.UpdateManagerService,
		InitFunc: registryInit,
	})
}

func newUpdateOrchestrator(eventsManager events.UpdateEventsManager, selfUpdateManager orchestration.UpdateManager, k8sOrchestrationManager orchestration.UpdateManager,
	rebootManager RebootManager, pahoClient mqtt.Client, cfg *mgrOpts) *updateOrchestrator {
	return &updateOrchestrator{
		cfg:                     cfg,
		rebootManager:           rebootManager,
		eventsManager:           eventsManager,
		selfUpdateManager:       selfUpdateManager,
		k8sOrchestrationManager: k8sOrchestrationManager,
		pahoClient:              pahoClient,
	}
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {

	orchServiceConfig, ok := registryCtx.Config.(map[string]interface{})
	if !ok {
		return nil, log.NewErrorf("incompatible configuration provided: %s", registryCtx.Config)
	}

	k8sMgrInitOpts, ok := orchServiceConfig["k8s"].([]k8s.MgrOpt)
	if !ok {
		return nil, log.NewErrorf("incompatible configuration provided: %s", orchServiceConfig["k8s"])
	}

	suMgrInitOpts, ok := orchServiceConfig["self_update"].([]selfupdate.MgrOpt)
	if !ok {
		return nil, log.NewErrorf("incompatible configuration provided: %s", orchServiceConfig["self_update"])
	}
	updOrchInitOpts, ok := orchServiceConfig["update_orchestrator"].([]MgrOpt)
	if !ok {
		return nil, log.NewErrorf("incompatible configuration provided: %s", orchServiceConfig["update_orchestrator"])
	}

	var (
		cfg = &mgrOpts{}
	)
	applyOptsMgr(cfg, updOrchInitOpts...)

	eventsManagerService, err := registryCtx.Get(registry.EventsManagerService)
	if err != nil {
		return nil, err
	}

	k8sOrchestrationManager, err := k8s.NewK8sUpdateManager(k8sMgrInitOpts, registryCtx)
	if err != nil {
		return nil, err
	}

	selfUpdateManager, err := selfupdate.NewSelfUpdateManager(suMgrInitOpts, registryCtx)
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

	rebootManager := &rebootManager{}

	//initialize the manager local service
	updOrch := newUpdateOrchestrator(eventsManagerService.(events.UpdateEventsManager), selfUpdateManager, k8sOrchestrationManager, rebootManager, pahoClient, cfg)

	if err := updOrch.subscribeRemoteConnectionStatus(); err != nil {
		return nil, err
	}

	return updOrch, nil
}

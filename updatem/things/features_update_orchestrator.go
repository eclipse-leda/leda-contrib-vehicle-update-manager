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
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Constants the define the respective Ditto feature based on the UpdateOrchestrator Vorto model
const (
	UpdateOrchestratorFeatureID                         = "UpdateOrchestrator"
	updateOrchestratorFeaturePropertyStatus             = "status"
	updateOrchestratorFeaturePropertyStatusState        = updateOrchestratorFeaturePropertyStatus + "/state"
	updateOrchestratorFeaturePropertyStatusCurrentState = updateOrchestratorFeaturePropertyStatus + "/currentState"
	updateOrchestratorFeatureOperationApply             = "apply"
)

var (
	defaultCurrentStateDelay = 30 * time.Second
)

type updateOrchestratorFeatureStatus struct {
	State        *manifestState               `json:"state"`
	CurrentState []*unstructured.Unstructured `json:"currentState"`
}

type updateOrchestratorFeature struct {
	status                *updateOrchestratorFeatureStatus
	orchMgr               orchestration.UpdateManager
	eventsMgr             events.UpdateEventsManager
	rootThing             model.Thing
	cancelEventsHandler   context.CancelFunc
	eventsHandlingLock    sync.Mutex
	processOperationsLock sync.Mutex
	updatesLock           sync.Mutex
	currentStateTimer     *time.Timer
}

func newUpdateOrchestratorFeature(rootThing model.Thing, eventsMgr events.UpdateEventsManager, orchMgr orchestration.UpdateManager) managedFeature {
	return &updateOrchestratorFeature{
		rootThing: rootThing,
		orchMgr:   orchMgr,
		eventsMgr: eventsMgr,
	}
}

func (updOrchFeature *updateOrchestratorFeature) register(ctx context.Context) error {
	log.Debug("initializing UpdateOrchestrator feature")
	if updOrchFeature.cancelEventsHandler == nil {
		updOrchFeature.handleEvents(ctx)
		log.Debug("subscribed for update events")
	}
	if err := updOrchFeature.rootThing.SetFeature(UpdateOrchestratorFeatureID, updOrchFeature.createFeature()); err != nil {
		return err
	}
	updOrchFeature.updateCurrentState(ctx)
	return nil
}

func (updOrchFeature *updateOrchestratorFeature) dispose() {
	log.Debug("disposing UpdateOrchestrator feature")
	if updOrchFeature.cancelEventsHandler != nil {
		log.Debug("unsubscribing from update manager events")
		updOrchFeature.cancelEventsHandler()
		updOrchFeature.cancelEventsHandler = nil
	}
}

func (updOrchFeature *updateOrchestratorFeature) featureOperationsHandler(operationName string, args interface{}) (interface{}, error) {
	ctx := context.Background()
	if operationName == updateOrchestratorFeatureOperationApply {
		log.Debug("received orchestrator manifest apply command")
		var ok bool
		var argsMap map[string]interface{}
		if argsMap, ok = args.(map[string]interface{}); !ok {
			return nil, client.NewMessagesParameterInvalidError("the parameter is not JSON object")
		}
		var correlationID string
		if correlationID, ok = argsMap["correlationId"].(string); !ok {
			return nil, client.NewMessagesParameterInvalidError("the correlation id is not string")
		}
		if correlationID == "" {
			return nil, client.NewMessagesParameterInvalidError("missing correlation id")
		}
		var yamlContent string
		if yamlContent, ok = argsMap["payload"].(string); !ok {
			return nil, client.NewMessagesParameterInvalidError("the YAML content is not string")
		}
		_, manifest, err := parseMultiYAML([]byte(yamlContent))
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		ctx := setApplyCorrelationIDContext(ctx, correlationID)
		return nil, updOrchFeature.apply(ctx, manifest)
	}
	err := log.NewErrorf("unsupported operation %s", operationName)
	log.ErrorErr(err, "unsupported operation %s", operationName)
	return nil, client.NewMessagesSubjectNotFound(err.Error())
}

func (updOrchFeature *updateOrchestratorFeature) apply(ctx context.Context, mf []*unstructured.Unstructured) error {
	go updOrchFeature.processApply(ctx, mf)
	return nil
}

func (updOrchFeature *updateOrchestratorFeature) processApply(ctx context.Context, mf []*unstructured.Unstructured) {
	updOrchFeature.processOperationsLock.Lock()
	defer updOrchFeature.processOperationsLock.Unlock()

	log.Debug("processing apply manifest command")
	updOrchFeature.orchMgr.Apply(ctx, mf)
	log.Debug("processing apply manifest command - done")
}

func (updOrchFeature *updateOrchestratorFeature) createFeature() model.Feature {
	return client.NewFeature(UpdateOrchestratorFeatureID,
		client.WithFeatureProperty(updateOrchestratorFeaturePropertyStatus, updOrchFeature.status),
		client.WithFeatureOperationsHandler(updOrchFeature.featureOperationsHandler),
	)
}

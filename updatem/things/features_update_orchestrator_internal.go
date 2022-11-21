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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (updOrchFeature *updateOrchestratorFeature) updateStatus(mfStatus manifestStatus, mfError *manifestError, correlationID string) {
	updOrchFeature.updatesLock.Lock()
	defer updOrchFeature.updatesLock.Unlock()
	if updOrchFeature.status == nil {
		log.Debug("no configured manifest - skipping update")
		return
	}
	updOrchFeature.status.State.Status = mfStatus
	updOrchFeature.status.State.Error = mfError
	updOrchFeature.status.State.CorrelationID = correlationID

	if err := updOrchFeature.rootThing.SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusState, updOrchFeature.status.State); err != nil {
		log.Error("could not update the UpdateOrchestrator feature status property: %v", err)
	}
}

func (updOrchFeature *updateOrchestratorFeature) updateCurrentState(ctx context.Context) {
	updOrchFeature.updatesLock.Lock()
	defer updOrchFeature.updatesLock.Unlock()
	if updOrchFeature.status == nil {
		updOrchFeature.status = &updateOrchestratorFeatureStatus{}
	}
	updOrchFeature.status.CurrentState = updOrchFeature.orchMgr.Get(ctx)
	if err := updOrchFeature.rootThing.SetFeatureProperty(UpdateOrchestratorFeatureID, updateOrchestratorFeaturePropertyStatusCurrentState, updOrchFeature.status.CurrentState); err != nil {
		log.Error("could not update the UpdateOrchestrator feature status/currentState property: %v", err)
	}
}

func (updOrchFeature *updateOrchestratorFeature) updateState(mf []*unstructured.Unstructured) {
	updOrchFeature.updatesLock.Lock()
	defer updOrchFeature.updatesLock.Unlock()

	updOrchFeature.status = &updateOrchestratorFeatureStatus{State: &manifestState{
		Manifest: mf,
	}}
}

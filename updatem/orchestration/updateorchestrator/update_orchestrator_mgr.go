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
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/selfupdate"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UpdateOrchestratorServiceID is the ID of the locally registered update orchestrator manager
const UpdateOrchestratorServiceID = "update.orchestrator.mgr"

type updateOrchestrator struct {
	applyLock               sync.Mutex
	cfg                     *mgrOpts
	rebootManager           RebootManager
	eventsManager           events.UpdateEventsManager
	selfUpdateManager       orchestration.UpdateManager
	k8sOrchestrationManager orchestration.UpdateManager
	connectionStatus        bool
	pahoClient              mqtt.Client
}

func (upOrch *updateOrchestrator) Apply(ctx context.Context, mf []*unstructured.Unstructured) interface{} {
	upOrch.applyLock.Lock()
	log.Debug("performing update operation...")

	var applyErr error
	var suApplyResult *selfupdate.ApplyResult
	applyCtx := orchestration.SetUpdateMgrApplyContext(ctx, mf)

	defer func() {
		upOrch.applyLock.Unlock()
	}()

	upOrch.publishOrchestrationEvent(applyCtx, orchestration.EventActionOrchestrationStarted, nil)

	var selfUpdateManifest *unstructured.Unstructured
	manifest := []*unstructured.Unstructured{}
	for i := 0; i < len(mf); i++ {
		if mf[i].GetKind() == "SelfUpdateBundle" {
			if selfUpdateManifest != nil {
				applyErr = log.NewErrorf("more than one SelfUpdateBundle resource in the YAML manifest")
				upOrch.publishOrchestrationEvent(applyCtx, orchestration.EventActionOrchestrationFinished, applyErr)
				return nil
			}
			selfUpdateManifest = mf[i]
		} else {
			manifest = append(manifest, mf[i])
		}
	}

	if selfUpdateManifest != nil {
		log.Debug("processing self update")
		suApplyResult = upOrch.selfUpdateManager.Apply(applyCtx, []*unstructured.Unstructured{selfUpdateManifest}).(*selfupdate.ApplyResult)
		applyErr = suApplyResult.Err
		log.Debug("processing self update - done")
	}

	if applyErr == nil && len(manifest) > 0 {
		log.Debug("processing apply manifest command")
		err := upOrch.k8sOrchestrationManager.Apply(applyCtx, manifest)
		if err != nil {
			applyErr = err.(error)
		}
		log.Debug("processing apply manifest command - done")
	}

	upOrch.publishOrchestrationEvent(applyCtx, orchestration.EventActionOrchestrationFinished, applyErr)

	if suApplyResult != nil && suApplyResult.RebootRequired {
		if err := upOrch.rebootManager.Reboot(suApplyResult.RebootTimeout); err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

func (upOrch *updateOrchestrator) Get(ctx context.Context) []*unstructured.Unstructured {
	result := []*unstructured.Unstructured{}
	result = append(result, upOrch.k8sOrchestrationManager.Get(ctx)...)
	result = append(result, upOrch.selfUpdateManager.Get(ctx)...)
	return result
}

func (upOrch *updateOrchestrator) Dispose(ctx context.Context) error {
	return nil
}

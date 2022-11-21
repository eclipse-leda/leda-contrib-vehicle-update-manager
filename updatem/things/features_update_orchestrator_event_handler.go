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
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (updOrchFeature *updateOrchestratorFeature) handleEvents(ctx context.Context) {
	subscribeCtx, subscrCtxCancelFunc := context.WithCancel(ctx)
	updOrchFeature.cancelEventsHandler = subscrCtxCancelFunc
	eventsChannel, errorChannel := updOrchFeature.eventsMgr.Subscribe(subscribeCtx)
	go func(ctx context.Context) error {
	eventsLoop:
		for {
			select {
			case evt := <-eventsChannel:
				switch evt.Type {
				case orchestration.EventTypeOrchestration:
					updOrchFeature.handleOrchestrationEvent(evt)
				case events.EventTypeResources:
					updOrchFeature.handleResourcesEvent(evt)
				default:
					log.Debug("event received that does not affect the UpdateOrchestrator feature")
				}
			case err := <-errorChannel:
				log.ErrorErr(err, "received Error from subscription")
			case <-ctx.Done():
				log.Debug("subscribe context is done - exiting subscribe events loop")
				break eventsLoop

			}
		}
		return nil
	}(subscribeCtx)
}

func (updOrchFeature *updateOrchestratorFeature) handleOrchestrationEvent(evt *events.Event) {
	switch evt.Action {
	case orchestration.EventActionOrchestrationStarted:
		updOrchFeature.handleOrchestrationStartedEvent(evt)
	case orchestration.EventActionOrchestrationRunning:
		updOrchFeature.handleOrchestrationRunningEvent(evt)
	case orchestration.EventActionOrchestrationFinished:
		updOrchFeature.handleOrchestrationFinishedEvent(evt)
	default:
		log.Debug("event received that does not affect the UpdateOrchestrator feature")
	}
}

func (updOrchFeature *updateOrchestratorFeature) handleResourcesEvent(evt *events.Event) {
	switch evt.Action {
	case events.EventActionResourcesAdded,
		events.EventActionResourcesUpdated,
		events.EventActionResourcesDeleted:
		updOrchFeature.handleResourceEvent(evt)
	default:
		log.Debug("resource changed event received that does not affect the UpdateOrchestrator feature")
	}
}

func (updOrchFeature *updateOrchestratorFeature) handleOrchestrationStartedEvent(event *events.Event) {
	updOrchFeature.eventsHandlingLock.Lock()
	defer updOrchFeature.eventsHandlingLock.Unlock()
	correlationID := getApplyCorrelationIDContext(event.Context)
	updOrchFeature.updateState(orchestration.GetUpdateMgrApplyContext(event.Context))
	updOrchFeature.updateStatus(manifestStatusStarted, nil, correlationID)
}

func (updOrchFeature *updateOrchestratorFeature) handleOrchestrationRunningEvent(event *events.Event) {
	updOrchFeature.eventsHandlingLock.Lock()
	defer updOrchFeature.eventsHandlingLock.Unlock()
	correlationID := getApplyCorrelationIDContext(event.Context)
	updOrchFeature.updateStatus(manifestStatusRunning, nil, correlationID)
}
func (updOrchFeature *updateOrchestratorFeature) handleOrchestrationFinishedEvent(event *events.Event) {
	updOrchFeature.eventsHandlingLock.Lock()
	defer updOrchFeature.eventsHandlingLock.Unlock()
	correlationID := getApplyCorrelationIDContext(event.Context)
	if event.Error != nil {
		updOrchFeature.updateStatus(manifestStatusFinishedError, &manifestError{
			Code:    500,
			Message: event.Error.Error(),
		}, correlationID)
	} else {
		updOrchFeature.updateStatus(manifestStatusFinishedSuccess, nil, correlationID)
	}
	updOrchFeature.updateCurrentState(event.Context)
}
func (updOrchFeature *updateOrchestratorFeature) handleResourceEvent(event *events.Event) {
	updOrchFeature.eventsHandlingLock.Lock()
	defer updOrchFeature.eventsHandlingLock.Unlock()
	if updOrchFeature.currentStateTimer != nil {
		updOrchFeature.currentStateTimer.Stop()
	}
	flagUpdateState := false
	if event.Source != nil {
		eventSource, ok := event.Source.(*unstructured.Unstructured)
		if ok && eventSource.GetKind() == "SelfUpdateBundle" {
			flagUpdateState = true
		}
		_, ok = event.Source.([]*unstructured.Unstructured)
		if ok {
			flagUpdateState = true
		}
	}
	if flagUpdateState {
		updOrchFeature.updateCurrentState(event.Context)
		return
	}
	updOrchFeature.currentStateTimer = time.AfterFunc(defaultCurrentStateDelay, func() {
		updOrchFeature.updateCurrentState(event.Context)
	})
}

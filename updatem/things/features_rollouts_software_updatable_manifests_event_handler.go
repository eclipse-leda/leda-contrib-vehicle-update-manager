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
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
)

func (suMf *softwareUpdatableManifests) handleOrchestrationEvents(ctx context.Context) {
	subscribeCtx, subscrCtxCancelFunc := context.WithCancel(ctx)
	suMf.cancelEventsHandler = subscrCtxCancelFunc

	eventsChannel, errorChannel := suMf.eventsMgr.Subscribe(subscribeCtx)
	go func(ctx context.Context) error {
	eventsLoop:
		for {
			select {
			case evt := <-eventsChannel:
				if evt.Type == orchestration.EventTypeOrchestration {
					if suMf.validateOperationContext(evt.Context) {
						switch evt.Action {
						case orchestration.EventActionOrchestrationStarted:
							suMf.handleEventStarted(evt)
						case orchestration.EventActionOrchestrationFinished:
							suMf.handleEventFinished(evt)
						default:
							log.Debug("an event that is not related to SoftwareUpdatable:manifest status reporting has been received")
						}
					} else {
						log.Debug("the event context is not relevant to the current operation status - will not process it")
					}
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

func (suMf *softwareUpdatableManifests) handleEventStarted(event *events.Event) {
	ctxOpStatus := getSUInstallContext(event.Context)
	ctxOpStatus.Status = datatypes.Installing
	suMf.updateLastOperation(ctxOpStatus)
}

func (suMf *softwareUpdatableManifests) handleEventFinished(event *events.Event) {
	log.Debug("got finished event - start processing")
	ctxOpStatus := getSUInstallContext(event.Context)
	if event.Error != nil {
		log.Debug("last operation has failed - will update lastFailedOperation property")
		ctxOpStatus.Message = event.Error.Error()
		ctxOpStatus.Status = datatypes.FinishedError
		suMf.updateLastFailedOperation(ctxOpStatus)
	} else {
		log.Debug("last operation has succeeded - will update lastOperation property to Installed")
		ctxOpStatus.Status = datatypes.Installed
		suMf.updateLastOperation(ctxOpStatus)
		ctxOpStatus.Status = datatypes.FinishedSuccess
	}
	suMf.updateLastOperation(ctxOpStatus)
}

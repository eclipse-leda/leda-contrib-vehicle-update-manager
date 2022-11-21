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

package events

import (
	"context"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
)

// EventsManagerServiceLocalID is the ID used by the service in the local services registry
const EventsManagerServiceLocalID = "updatem.service.local.v1.service-events-manager"

func init() {
	registry.Register(&registry.Registration{
		ID:       EventsManagerServiceLocalID,
		Type:     registry.EventsManagerService,
		InitFunc: registryInit,
	})
}

type eventsMgr struct {
	broadcaster  *eventsSinkDispatcher
	publishMutex sync.Mutex
}

func (eMgr *eventsMgr) Publish(ctx context.Context, event *Event) error {
	eMgr.publishMutex.Lock()
	defer eMgr.publishMutex.Unlock()

	err := eMgr.broadcaster.write(event)
	if err != nil {
		log.ErrorErr(err, "could not publish event: %+v", event)
	}
	//log.Trace("published event %+v", event)
	return err
}

func (eMgr *eventsMgr) Subscribe(ctx context.Context) (<-chan *Event, <-chan error) {
	var (
		eventsEmitter               = make(chan *Event)
		errorsEmitter               = make(chan error, 1)
		broadcasterChan             = newChannelledEventsSink(0)
		broadcasterQueue            = newQueueEventsSink(broadcasterChan)
		resultsSink      eventsSink = broadcasterQueue
	)

	clearResources := func() {
		close(errorsEmitter)
		eMgr.broadcaster.remove(resultsSink)
		broadcasterQueue.close()
		broadcasterChan.close()
	}

	eMgr.broadcaster.add(resultsSink)

	go func() {
		defer clearResources()

		var err error
	eventsLoop:
		for {
			select {
			case internalEvent := <-broadcasterChan.eventsChannel:
				event, ok := internalEvent.(*Event)
				if !ok {
					err = log.NewErrorf("invalid message received: %#v", internalEvent)
					log.DebugErr(err, "invalid message received")
					break
				}
				select {
				case eventsEmitter <- event:
					//log.Trace("sent event to subscriber %+v", event)
				case <-ctx.Done():
					log.Debug("subscriber context is done")
					break eventsLoop
				}
			case <-ctx.Done():
				log.Debug("subscriber context is done")
				break eventsLoop
			}
		}
		if err == nil {
			if ctxErr := ctx.Err(); ctxErr != context.Canceled {
				log.DebugErr(ctxErr, "subscriber context has an error")
				err = ctxErr
			}
		}
		errorsEmitter <- err
	}()

	return eventsEmitter, errorsEmitter
}

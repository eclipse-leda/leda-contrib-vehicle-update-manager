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
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
)

type eventsManagerPrep func() UpdateEventsManager

func TestPublishErr(t *testing.T) {
	var ctx context.Context
	pubErrTests := map[string]struct {
		eventsMgrInit eventsManagerPrep
		event         *Event
		err           error
	}{
		"test_sink_closed": {
			eventsMgrInit: func() UpdateEventsManager {
				broadcaster := newEventSinksDispatcher()
				broadcaster.close()
				return &eventsMgr{broadcaster: broadcaster}
			},
			event: &Event{
				Type:   EventTypeResources,
				Action: EventActionResourcesAdded,
			},
			err: errEventsSinkClosed,
		},
	}

	for testName, tc := range pubErrTests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			var evMgr UpdateEventsManager
			if tc.eventsMgrInit != nil {
				evMgr = tc.eventsMgrInit()
			} else {
				evMgr = newEventsManager()
			}
			err := evMgr.Publish(ctx, tc.event)
			testutil.AssertError(t, tc.err, err)
		})
	}
}

func TestSubscribe(t *testing.T) {
	evMgr := newEventsManager()

	ctx := context.Background()

	subscribeCtx, subscribeCtxCancelFunc := context.WithCancel(ctx)
	t.Cleanup(subscribeCtxCancelFunc)
	eventsChan, eventsErrChan := evMgr.Subscribe(subscribeCtx)

	testEvents := []*Event{
		{
			Type:   EventTypeResources,
			Action: EventActionResourcesAdded,
		},
		{
			Type:   EventTypeResources,
			Action: EventActionResourcesUpdated,
		},
	}

	t.Log("publish test events ")
	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go func() {
		defer wg.Done()
		defer close(errChan)

		for _, testEvent := range testEvents {
			if err := evMgr.Publish(ctx, testEvent); err != nil {
				errChan <- err
				return
			}
		}
		t.Log("finished publishing test events")
	}()

	t.Log("waiting to publish all test events")
	wg.Wait()
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
	var received []*Event
assertSubscribe:
	for {
		select {
		case msg := <-eventsChan:
			received = append(received, msg)
		case err := <-eventsErrChan:
			if err != nil {
				t.Errorf("unexpected error received: %v", err)
				t.Fatal(err)
			}
			break assertSubscribe
		}
		if len(received) == len(testEvents) {
			subscribeCtxCancelFunc()
			for i, receivedEvent := range received {
				testutil.AssertEqual(t, testEvents[i], receivedEvent)
			}
		}
	}
}

func TestSubscribeContextError(t *testing.T) {
	evMgr := newEventsManager()
	subscribeCtx, subscribeCancelFunc := context.WithDeadline(context.Background(), time.Now().UTC())
	t.Cleanup(subscribeCancelFunc)
	eventsChan, eventsErrChan := evMgr.Subscribe(subscribeCtx)

assertSubscribe:
	for {
		select {
		case msg := <-eventsChan:
			if msg != nil {
				t.Errorf("unexpected message received: %v", msg)
				t.Fatal(msg)
			}
			break assertSubscribe
		case err := <-eventsErrChan:
			testutil.AssertEqual(t, context.DeadlineExceeded, err)
			break assertSubscribe
		}
	}
}

func TestSubscribeInvalidMessageError(t *testing.T) {
	broadcaster := newEventSinksDispatcher()
	evMgr := &eventsMgr{
		broadcaster: broadcaster,
	}

	subscribeCtx, subscribeCtxCancelFunc := context.WithCancel(context.Background())
	eventsChan, eventsErrChan := evMgr.Subscribe(subscribeCtx)

	expectedErr := log.NewError("test expected error")

	t.Log("publish test events ")
	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go func() {
		defer wg.Done()
		defer close(errChan)
		if err := broadcaster.write(expectedErr); err != nil {
			errChan <- err
			return
		}
		t.Log("finished publishing test events")
	}()

	t.Log("waiting to publish all test events")
	wg.Wait()
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(1 * time.Second)
		subscribeCtxCancelFunc()
	}()

assertSubscribe:
	for {
		select {
		case msg := <-eventsChan:
			if msg != nil {
				t.Errorf("unexpected message received: %v", msg)
				t.Fatal(msg)
			}
			break assertSubscribe
		case err := <-eventsErrChan:
			testutil.AssertError(t, log.NewErrorf("invalid message received: %#v", expectedErr), err)
			break assertSubscribe
		}
	}
}

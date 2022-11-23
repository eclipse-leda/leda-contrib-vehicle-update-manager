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
	"encoding/json"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	topicRemoteStatus = "edge/connection/remote/status"
)

func (updOrch *updateOrchestrator) publishEvent(ctx context.Context, eventType events.EventType, eventAction events.EventAction, eventSource []*unstructured.Unstructured, err error) {
	e := &events.Event{
		Type:    eventType,
		Action:  eventAction,
		Time:    time.Now().UTC().Unix(),
		Source:  eventSource,
		Context: ctx,
		Error:   err,
	}

	if pubErr := updOrch.eventsManager.Publish(ctx, e); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event [%+v]", e)
	}
}

func (updOrch *updateOrchestrator) publishOrchestrationEvent(ctx context.Context, eventAction events.EventAction, err error) {
	updOrch.publishEvent(ctx, orchestration.EventTypeOrchestration, eventAction, nil, err)
}

func (updOrch *updateOrchestrator) publishResourceEvent(ctx context.Context, eventAction events.EventAction, eventSource []*unstructured.Unstructured, err error) {
	updOrch.publishEvent(ctx, events.EventTypeResources, eventAction, eventSource, err)
}

func (updOrch *updateOrchestrator) handleConnectionStatus(mqttClient mqtt.Client, message mqtt.Message) {
	log.Debug("received remote connection status = %s", string(message.Payload()))

	connectionStatus := map[string]interface{}{}

	err := json.Unmarshal(message.Payload(), &connectionStatus)
	if err != nil {
		log.Error("error unmarshal message payload to connectionStatus = %s \n", err)
		return
	}
	connStatus, ok := connectionStatus["connected"].(bool)
	if !ok {
		log.Error("error missing or invalid 'connected' property in connnection status payload")
		return
	}

	if connStatus && !updOrch.connectionStatus {
		u := updOrch.Get(context.Background())
		updOrch.publishResourceEvent(context.Background(), events.EventActionResourcesUpdated, u, nil)
	}
	updOrch.connectionStatus = connStatus
}

func (updOrch *updateOrchestrator) subscribeRemoteConnectionStatus() error {
	log.Debug("subscribing for '%s' topic", topicRemoteStatus)
	if token := updOrch.pahoClient.Subscribe(topicRemoteStatus, 1, updOrch.handleConnectionStatus); !token.WaitTimeout(updOrch.cfg.acknowledgeTimeout) {
		return log.NewErrorf("cannot subscribe for topic '%s' in '%v' seconds", topicRemoteStatus, updOrch.cfg.acknowledgeTimeout)
	}
	return nil
}

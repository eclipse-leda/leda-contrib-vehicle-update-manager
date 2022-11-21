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
	"context"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func (suMgr *selfUpdateManager) publishEvent(ctx context.Context, eventType events.EventType, eventAction events.EventAction, eventSource *unstructured.Unstructured, err error) {
	e := &events.Event{
		Type:    eventType,
		Action:  eventAction,
		Time:    time.Now().UTC().Unix(),
		Source:  eventSource,
		Context: ctx,
		Error:   err,
	}

	if pubErr := suMgr.eventsMgr.Publish(ctx, e); pubErr != nil {
		log.ErrorErr(pubErr, "failed to publish event [%+v]", e)
	}
}

func (suMgr *selfUpdateManager) publishResourceEvent(ctx context.Context, eventAction events.EventAction, eventSource *unstructured.Unstructured, err error) {
	suMgr.publishEvent(ctx, events.EventTypeResources, eventAction, eventSource, err)
}

func (suMgr *selfUpdateManager) subscribeSelfUpdateCurrentState() error {
	log.Debug("subscribing for '%s' topic", topicSelfUpdateCurrentState)
	if token := suMgr.pahoClient.Subscribe(topicSelfUpdateCurrentState, 1, suMgr.handleSelfUpdateCurrentState); !token.WaitTimeout(suMgr.cfg.acknowledgeTimeout) {
		return log.NewErrorf("cannot subscribe for topic '%s' in '%v' seconds", topicSelfUpdateCurrentState, suMgr.cfg.acknowledgeTimeout)
	}
	return nil
}

func (suMgr *selfUpdateManager) handleSelfUpdateCurrentState(mqttClient mqtt.Client, message mqtt.Message) {
	log.Debug("received self update current state")
	_, u, err := parseMultiYAML([]byte(message.Payload()))
	if err != nil {
		log.Error("error while parsing self update current state : %v", err)
		return
	}
	suMgr.currentState = u[0]
	suMgr.publishResourceEvent(context.Background(), events.EventActionResourcesUpdated, u[0], nil)
}

func (suMgr *selfUpdateManager) unmarshalUnstructured(u *unstructured.Unstructured) ([]byte, error) {
	jsonBytes, err := u.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := u.UnmarshalJSON(jsonBytes); err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(jsonBytes)
}

func (suMgr *selfUpdateManager) convertStringToDuration(value string, defaultValue time.Duration) time.Duration {
	durationValue, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return durationValue
}

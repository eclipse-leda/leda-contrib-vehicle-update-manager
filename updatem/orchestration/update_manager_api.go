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

package orchestration

import (
	"context"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// EventTypeOrchestration is the type of events that are related to orchestration
	EventTypeOrchestration events.EventType = "orchestration"

	// EventActionOrchestrationStarted is emitted each time an orchestration process is started
	EventActionOrchestrationStarted events.EventAction = "started"
	// EventActionOrchestrationRunning is emitted each time an orchestration process is running
	EventActionOrchestrationRunning events.EventAction = "running"
	// EventActionOrchestrationFinished is emitted each time an orchestration process has finished
	EventActionOrchestrationFinished events.EventAction = "finished"
)

// UpdateManager provides the orchestration management abstraction
type UpdateManager interface {
	Apply(ctx context.Context, mf []*unstructured.Unstructured) interface{}
	Get(ctx context.Context) []*unstructured.Unstructured
	Dispose(ctx context.Context) error
}

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

import "context"

// EventType represents the event's type
type EventType string

const (
	// EventTypeResources is an event type for the resources
	EventTypeResources EventType = "resources"
	// in the future more types will be added - e.g. for image changes, etc.
)

// EventAction represents the event's action
type EventAction string

const (
	// EventActionResourcesAdded is used when a Pod or Node resource is added
	EventActionResourcesAdded EventAction = "added"
	// EventActionResourcesUpdated is used when a Pod or Node resource is updated
	EventActionResourcesUpdated EventAction = "updated"
	// EventActionResourcesDeleted is used when a Pod or Node resource is deleted
	EventActionResourcesDeleted EventAction = "deleted"
)

// Event represents an emitted event
type Event struct {
	// the EventType
	Type EventType `json:"type"`
	// the EventAction
	Action EventAction `json:"action"`
	// the instance that changed
	Source interface{} `json:"source,omitempty"`
	// time
	Time int64 `json:"time,omitempty"`
	// event context
	Context context.Context `json:"context"`
	// error information about the event
	Error error `json:"error"`
}

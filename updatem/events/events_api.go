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
)

// UpdateEventsManager provies a simple way of publishing and subscribing to update manager related events.
type UpdateEventsManager interface {
	// Publish adds a new event to be dispatched based on the provided EventType and EventAction
	Publish(ctx context.Context, event *Event) error
	// Subscribe provides two channels where the according events and errors can be received via the subscriber context provided
	Subscribe(ctx context.Context) (<-chan *Event, <-chan error)
}

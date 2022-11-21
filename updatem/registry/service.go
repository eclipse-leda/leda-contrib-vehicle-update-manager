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

package registry

import "github.com/eclipse-kanto/container-management/containerm/registry"

const (
	// EventsManagerService implements THE events manager service
	EventsManagerService registry.Type = "updatemanagerd.service.events.manager.v1"
	// UpdateManagerService implements THE update manager service
	UpdateManagerService registry.Type = "updatemanagerd.service.update.manager.v1"
	// ThingsUpdateManagerService implements THE update orchestration via Rollouts and Things service
	ThingsUpdateManagerService registry.Type = "updatemanagerd.service.things.update.manager.v1"
)

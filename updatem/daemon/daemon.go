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

package main

import (
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
)

type daemon struct {
	config         *config
	serviceInfoSet *registry.Set
}

func newDaemon(config *config) (*daemon, error) {
	log.Debug("starting Update Manager daemon initialization")
	daemon := &daemon{
		config:         config,
		serviceInfoSet: registry.NewServiceInfoSet(),
	}
	log.Debug("successfully created Update Manager daemon instance")
	return daemon, nil
}

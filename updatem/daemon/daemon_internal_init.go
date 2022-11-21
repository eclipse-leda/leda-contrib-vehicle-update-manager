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
	"context"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	registryservices "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/registry"
)

func (d *daemon) init() {
	log.Debug("will start Update Manager initialization")

	registrationsMap := registry.RegistrationsMap()

	log.Debug("the current registered services ready for initialization are %+v", registrationsMap)

	var ctx = context.TODO()

	//init events manager services
	initService(ctx, d, registrationsMap, registry.EventsManagerService)

	//init update manager service
	initService(ctx, d, registrationsMap, registryservices.UpdateManagerService)

	//init Things update manager service
	initService(ctx, d, registrationsMap, registryservices.ThingsUpdateManagerService)
}

func initService(ctx context.Context, d *daemon, registrationsMap map[registry.Type][]*registry.Registration, regType registry.Type) {
	var config interface{}
	log.Debug("will initialize all %s services", regType)
	serviceRegs, ok := registrationsMap[regType]
	if ok {
		switch regType {
		case registryservices.UpdateManagerService:
			config = extractUpdateManagerOptions(d.config)
			break
		case registryservices.ThingsUpdateManagerService:
			config = extractThingsOptions(d.config)
			break
		default:
			config = nil
		}
		d.initServices(ctx, serviceRegs, config)
	} else {
		log.Debug("there are no %s services registered", regType)
	}
}

func (d *daemon) initServices(ctx context.Context, registrations []*registry.Registration, config interface{}) {
	var (
		regCtx   *registry.ServiceRegistryContext
		servInfo *registry.ServiceInfo
	)
	for _, reg := range registrations {
		regCtx = registry.NewContext(ctx, config, reg, d.serviceInfoSet)

		log.Debug("will initialize service instance with ID = %s with context %v", reg.ID, regCtx)
		servInfo = reg.Init(regCtx)
		if servInfo.Err() != nil {
			log.ErrorErr(servInfo.Err(), "error initializing service %s - will not add it to the local service registry", servInfo.Registration.Type)
			continue
		}
		log.Debug("successfully initialized service instance with ID = %s ", reg.ID)
		d.serviceInfoSet.Add(servInfo)
		log.Debug("successfully added service instance with ID = %s to the local service registry", reg.ID)
	}
}

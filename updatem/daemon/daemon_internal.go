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

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"

	"github.com/eclipse-kanto/container-management/containerm/log"
	registryservices "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/registry"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/things"
)

func (d *daemon) start() error {
	log.Debug("starting daemon instance")
	err := d.startThingsManagers()
	if err != nil {
		log.ErrorErr(err, "could not start the Things Update Manager Services")
		return err
	}
	return nil

}

func (d *daemon) stop() {
	log.Debug("stopping of the Update Manager daemon is requested and started")

	log.Debug("stopping management local services")
	d.stopUpdateManagers()

	log.Debug("stopping Things Update Manager service")
	d.stopThingsManagers()

	log.Debug("stopping of the Update Manager daemon finished")
}

func (d *daemon) startThingsManagers() error {
	log.Debug("starting Things Update Manager services ")
	grpcServerInfos := d.serviceInfoSet.GetAll(registryservices.ThingsUpdateManagerService)
	var (
		instnace interface{}
		err      error
	)

	log.Debug("there are %d Things Update Manager services to be started", len(grpcServerInfos))
	for _, servInfo := range grpcServerInfos {
		log.Debug("will try to start Things Update Manager service local service with ID = %s", servInfo.Registration.ID)
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Things Update Manager service instance - local service ID = %s ", servInfo.Registration.ID)
		} else {
			err = instnace.(things.UpdateThingsManager).Connect()
			if err != nil {
				log.ErrorErr(err, "could not start Things Update Manager service with service ID = %s ", servInfo.Registration.ID)
			} else {
				log.Debug("successfully started Things Update Manager service with service ID = %s ", servInfo.Registration.ID)
			}
		}
	}
	return err
}

func (d *daemon) stopThingsManagers() {
	log.Debug("will stop Things Update Manager services")
	grpcServerInfos := d.serviceInfoSet.GetAll(registryservices.ThingsUpdateManagerService)
	var (
		instnace interface{}
		err      error
	)

	for _, servInfo := range grpcServerInfos {
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Things Update Manager service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			instnace.(things.UpdateThingsManager).Disconnect()
			log.Debug("successfully stopped Things Update Manager service with service ID = %s ", servInfo.Registration.ID)
		}
	}
}

func (d *daemon) stopUpdateManagers() {
	log.Debug("will stop update management local services")
	updMrgServices := d.serviceInfoSet.GetAll(registryservices.UpdateManagerService)
	var (
		instnace interface{}
		err      error
	)
	log.Debug("there are %d update manager services to be stopped", len(updMrgServices))
	for _, servInfo := range updMrgServices {
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get update manager service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			ctx := context.Background()
			err = instnace.(orchestration.UpdateManager).Dispose(ctx)
			if err != nil {
				log.ErrorErr(err, "could not stop update manager service for service ID = %s", servInfo.Registration.ID)
			}
		}
	}
}

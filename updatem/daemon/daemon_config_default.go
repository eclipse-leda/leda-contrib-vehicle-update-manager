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
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/things"
)

const (
	// default daemon config
	daemonConfigFileDefault = "/etc/updatemanagerd/updatemanagerd-config.json"

	// default daemon debug config
	daemonDebugSysLogDefault = false

	// default things connection config
	thingsEnableDefault                      = true
	thingsMetaPathDefault                    = "/var/lib/updatemanagerd"
	thingsConnectionBrokerURLDefault         = "tcp://localhost:1883"
	thingsConnectionKeepAliveDefault         = 20000
	thingsConnectionDisconnectTimeoutDefault = 250
	thingsConnectionClientUsername           = ""
	thingsConnectionClientPassword           = ""
	thingsConnectTimeoutTimeoutDefault       = 30000
	thingsAcknowledgeTimeoutDefault          = 15000
	thingsSubscribeTimeoutDefault            = 15000
	thingsUnsubscribeTimeoutDefault          = 5000

	// default log config
	logFileDefault         = "log/update-manager.log"
	logLevelDefault        = "INFO"
	logFileSizeDefault     = 2
	logFileCountDefault    = 5
	logFileMaxAgeDefault   = 28
	logEnableSyslogDefault = false

	// default k8s config
	k8sKubeconfigDefault          = "" //etc/rancher/k3s/k3s.yaml
	k8sCreateUpdateTimeoutDefault = 20
	k8sDeleteTimeoutDefault       = 20

	// default self update config
	selfUpdateTimeoutDefault       = "10m"
	selfUpdateRebootTimeoutDefault = "30s"
	selfUpdateEnableRebootDefault  = false
)

var (
	// default things service features config
	thingsServiceFeaturesDefault = []string{things.SoftwareUpdatableManifestsFeatureID}
)

func getDefaultInstance() *config {
	return &config{
		Log: &log.Config{
			LogFile:       logFileDefault,
			LogLevel:      logLevelDefault,
			LogFileSize:   logFileSizeDefault,
			LogFileCount:  logFileCountDefault,
			LogFileMaxAge: logFileMaxAgeDefault,
			Syslog:        logEnableSyslogDefault,
		},
		ThingsConfig: &thingsConfig{
			ThingsMetaPath: thingsMetaPathDefault,
			Features:       thingsServiceFeaturesDefault,
			ThingsConnectionConfig: &thingsConnectionConfig{
				BrokerURL:          thingsConnectionBrokerURLDefault,
				KeepAlive:          thingsConnectionKeepAliveDefault,
				DisconnectTimeout:  thingsConnectionDisconnectTimeoutDefault,
				ClientUsername:     thingsConnectionClientUsername,
				ClientPassword:     thingsConnectionClientPassword,
				ConnectTimeout:     thingsConnectTimeoutTimeoutDefault,
				AcknowledgeTimeout: thingsAcknowledgeTimeoutDefault,
				SubscribeTimeout:   thingsSubscribeTimeoutDefault,
				UnsubscribeTimeout: thingsUnsubscribeTimeoutDefault,
			},
		},
		Orchestration: &orchestrationConfig{
			K8s: &k8sExecutionConfig{
				Kubeconfig: k8sKubeconfigDefault,
			},
			SelfUpdate: &selfUpdateExecutionConfig{
				Timeout:       selfUpdateTimeoutDefault,
				RebootTimeout: selfUpdateRebootTimeoutDefault,
				EnableReboot:  selfUpdateEnableRebootDefault,
			},
		},
	}
}

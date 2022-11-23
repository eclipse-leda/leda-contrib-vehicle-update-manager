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
)

// config refers to daemon's whole configurations.
type config struct {
	Log           *log.Config          `json:"log,omitempty"`
	ThingsConfig  *thingsConfig        `json:"things,omitempty"`
	Orchestration *orchestrationConfig `json:"orchestration,omitempty"`
}

// things client configuration
type thingsConfig struct {
	ThingsMetaPath         string                  `json:"home_dir,omitempty"`
	Features               []string                `json:"features,omitempty"`
	ThingsConnectionConfig *thingsConnectionConfig `json:"connection,omitempty"`
}

// things service connection config
type thingsConnectionConfig struct {
	BrokerURL          string `json:"broker_url,omitempty"`
	KeepAlive          int64  `json:"keep_alive,omitempty"`
	DisconnectTimeout  int64  `json:"disconnect_timeout,omitempty"`
	ClientUsername     string `json:"client_username,omitempty"`
	ClientPassword     string `json:"client_password,omitempty"`
	ConnectTimeout     int64  `json:"connect_timeout,omitempty"`
	AcknowledgeTimeout int64  `json:"acknowledge_timeout,omitempty"`
	SubscribeTimeout   int64  `json:"subscribe_timeout,omitempty"`
	UnsubscribeTimeout int64  `json:"unsubscribe_timeout,omitempty"`
}

// k8s execution config
type k8sExecutionConfig struct {
	Kubeconfig string `json:"kubeconfig,omitempty"`
}

// self update executor config
type selfUpdateExecutionConfig struct {
	EnableReboot  bool   `json:"enable_reboot,omitempty"`
	Timeout       string `json:"timeout,omitempty"`
	RebootTimeout string `json:"reboot_timeout,omitempty"`
}

// orchestration config
type orchestrationConfig struct {
	K8s        *k8sExecutionConfig        `json:"k8s,omitempty"`
	SelfUpdate *selfUpdateExecutionConfig `json:"self_update,omitempty"`
}

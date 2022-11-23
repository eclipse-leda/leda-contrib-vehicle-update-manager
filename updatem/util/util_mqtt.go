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

package util

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenkalti/backoff/v3"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var backoffInterval = 5 * time.Second

// MqttConnect connects to the local client.
func MqttConnect(pahoClient mqtt.Client, broker string) error {
	b := backoff.NewConstantBackOff(backoffInterval)

	ticker := backoff.NewTicker(b)
	defer ticker.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	defer signal.Stop(sigs)

	for range ticker.C {
		future := pahoClient.Connect()

		select {
		case <-future.Done():
			err := future.Error()
			if err == nil {
				return nil
			}
		case <-sigs:
			return fmt.Errorf("connect to local broker on %s cancelled by signal", broker)
		}
	}

	return nil
}

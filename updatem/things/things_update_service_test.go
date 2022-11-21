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

package things

import (
	"net"
	"testing"
	"time"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/golang/mock/gomock"
)

// runs a dummy MQTT broker over TCP to assert the basic auth credentials in the MQTT Connect packet headers.
func TestThingsUpdateServiceConnectWithCredentials(t *testing.T) {
	const (
		testMQTTUsername  = "test-username"
		testMQTTPassword  = "test-passoword"
		testMQTTBrokerURL = "localhost:9999"
	)

	controller := gomock.NewController(t)
	defer func() {
		controller.Finish()
	}()
	setupEventsManagerMock(controller)
	setupThingsUpdateManager(controller)
	testThingsMgr = newThingsUpdateManager(mockUpdateManager, mockEventsManager,
		testMQTTBrokerURL,
		0,
		0,
		testMQTTUsername,
		testMQTTPassword,
		testThingsStoragePath,
		testThingsFeaturesDefaultSet,
		0,
		0,
		0,
		0)
	setupThingMock(controller)

	listener, err := net.Listen("tcp4", testMQTTBrokerURL)
	defer listener.Close()
	go func() {
		// wait the tcp listener to initialize
		time.Sleep(1 * time.Second)
		testThingsMgr.thingsClient.Connect()
	}()
	conn, err := listener.Accept()

	if err != nil {
		t.Errorf("Connection accept failure: %s", err)
	}
	controlPacket, err := packets.ReadPacket(conn)
	if err != nil {
		t.Errorf("reading err: %s", err)
	}

	connectPacket := controlPacket.(*packets.ConnectPacket)
	testutil.AssertEqual(t, testMQTTUsername, connectPacket.Username)
	testutil.AssertEqual(t, testMQTTPassword, string(connectPacket.Password))
}

// runs a dummy MQTT broker over TCP to assert the MQTT Connect packet headers.
func TestThingsUpdateServiceConnectNoCredentials(t *testing.T) {
	const (
		testMQTTBrokerURL = "localhost:9998"
	)

	controller := gomock.NewController(t)
	defer func() {
		controller.Finish()
	}()
	setupEventsManagerMock(controller)
	setupThingsUpdateManager(controller)
	testThingsMgr = newThingsUpdateManager(mockUpdateManager, mockEventsManager,
		testMQTTBrokerURL,
		0,
		0,
		"",
		"",
		testThingsStoragePath,
		testThingsFeaturesDefaultSet,
		0,
		0,
		0,
		0)
	setupThingMock(controller)

	listener, err := net.Listen("tcp4", testMQTTBrokerURL)
	defer listener.Close()
	go func() {
		// wait the tcp listener to initialize
		time.Sleep(1 * time.Second)
		testThingsMgr.thingsClient.Connect()
	}()
	conn, err := listener.Accept()

	if err != nil {
		t.Errorf("Connection accept failure: %s", err)
	}

	controlPacket, err := packets.ReadPacket(conn)
	if err != nil {
		t.Errorf("reading err: %s", err)
	}

	connectPacket := controlPacket.(*packets.ConnectPacket)
	testutil.AssertEqual(t, "", connectPacket.Username)
	testutil.AssertEqual(t, "", string(connectPacket.Password))
}

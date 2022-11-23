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

package selfupdate

import (
	"github.com/eclipse-kanto/container-management/containerm/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type selfUpdateOperation struct {
	done   chan bool
	err    error
	result OperationResult
}

// OperationResult holds the result of a self update operation
type OperationResult int

const (
	//SelfUpdateNoResult represents a self update operation initial state
	SelfUpdateNoResult OperationResult = iota
	// SelfUpdateResultInstalled represents a self update operation finished successfully
	SelfUpdateResultInstalled
	// SelfUpdateResultRejected represents a self rejected update operation
	SelfUpdateResultRejected
	// SelfUpdateResultError represents a self update operation failed with error
	SelfUpdateResultError
	// SelfUpdateResultTimeout represents a self update operation failed with expired timeout
	SelfUpdateResultTimeout
)

type selfUpdateState string

const (
	// selfUpdateStateIdle          selfUpdateState = "idle"
	selfUpdateStateFailed        selfUpdateState = "failed"
	selfUpdateStateInstalled     selfUpdateState = "installed"
	selfUpdateStateInstalling    selfUpdateState = "installing"
	selfUpdateStateDownloading   selfUpdateState = "downloading"
	selfUpdateStateUninitialized selfUpdateState = "uninitialized"
)

type stateTechCode int

const (
	techCodeDownloadFailed     stateTechCode = 1001
	techCodeInvalidBundle      stateTechCode = 2001
	techCodeInstallationFailed stateTechCode = 3001
	techCodeUpdateRejected     stateTechCode = 4001
	techCodeUnknownError       stateTechCode = 5001
)

var (
	stateTechCodeDescriptions = map[stateTechCode]string{
		techCodeDownloadFailed:     "Download failed",
		techCodeInvalidBundle:      "Invalid bundle",
		techCodeInstallationFailed: "Installation failed",
		techCodeUpdateRejected:     "Update rejected, bundle version same as current OS version",
		techCodeUnknownError:       "Unknown Error",
	}
)

func newSelfUpdateOperation() *selfUpdateOperation {
	return &selfUpdateOperation{
		done: make(chan bool, 1),
	}
}

func (su *selfUpdateOperation) handleSelfUpdateDesiredStateFeedback(mqttClient mqtt.Client, message mqtt.Message) {
	_, u, err := parseMultiYAML([]byte(message.Payload()))
	if err != nil {
		log.Error("error while parsing self update desiredstate feedback: %v", err)
		return
	}

	var ok bool
	var metadata map[string]interface{}
	if metadata, ok = su.getSelfUpdateFieldMap(u[0].Object, "metadata", "metadata"); !ok {
		return
	}
	var metadataName string
	if metadataName, ok = su.getSelfUpdateFieldString(metadata, "name", "metadata name"); !ok {
		return
	}
	var state map[string]interface{}
	if state, ok = su.getSelfUpdateFieldMap(u[0].Object, "state", "state"); !ok {
		return
	}
	var stateNameStr string
	if stateNameStr, ok = su.getSelfUpdateFieldString(state, "name", "state name"); !ok {
		return
	}
	stateName := selfUpdateState(stateNameStr)
	log.Debug("received self update desiredstate feedback for bundle name '%s' and state '%s'", metadataName, stateName)

	if stateName == selfUpdateStateUninitialized {
		su.err = log.NewErrorf("cannot perform self update operation. SUA is not configured yet")
		su.result = SelfUpdateResultError
		log.Error(su.err.Error())
		su.done <- true
	} else if stateName == selfUpdateStateDownloading || stateName == selfUpdateStateInstalling {
		var ok bool
		var progress int64
		if progress, ok = su.getSelfUpdateFieldInt64(state, "progress", "state progress"); !ok {
			return
		}
		if !su.checkStateProgress(progress) {
			log.Error("the self update state progress in the self update desiredstate feedback is invalid")
			return
		}
		if stateName == selfUpdateStateInstalling {
			log.Info("self update bundle is installing with progress '%v'", progress)
		} else {
			log.Info("self update bundle is downloading with progress '%v'", progress)
		}
	} else if stateName == selfUpdateStateFailed {
		var ok bool
		var techCodeInt int64
		if techCodeInt, ok = su.getSelfUpdateFieldInt64(state, "techCode", "state tech code"); !ok {
			return
		}
		techCode := stateTechCode(techCodeInt)
		if !su.isStateTechCodeSupported(techCode) {
			log.Error("the self update state tech code in the self update desiredstate feedback is not supported")
			return
		}
		if techCode == techCodeUpdateRejected {
			log.Info("the self update bundle is rejected, the bundle version is the same as current OS version")
			su.result = SelfUpdateResultRejected
			su.done <- true
			return
		}
		su.err = log.NewErrorf("self update operation is finished unsuccessfully with error : '%v'", stateTechCodeDescriptions[techCode])
		su.result = SelfUpdateResultError
		log.Error(su.err.Error())
		su.done <- true
	} else if stateName == selfUpdateStateInstalled {
		log.Info("the self update bundle is installed successfully")
		su.result = SelfUpdateResultInstalled
		su.done <- true
	}
}

func (su *selfUpdateOperation) checkStateProgress(progress int64) bool {
	return (progress >= 0) && (progress <= 100)
}

func (su *selfUpdateOperation) isStateTechCodeSupported(techCode stateTechCode) bool {
	if techCode == techCodeDownloadFailed || techCode == techCodeInstallationFailed || techCode == techCodeInvalidBundle ||
		techCode == techCodeUpdateRejected || techCode == techCodeUnknownError {
		return true
	}
	return false
}

func (su *selfUpdateOperation) getSelfUpdateFieldInt64(mapValue map[string]interface{}, mapField string, fieldName string) (int64, bool) {
	if value, ok := mapValue[mapField]; !ok {
		log.Error("missing self update %s in the self update desiredstate feedback", fieldName)
		return 0, false
	} else if convertedValue, ok := value.(int64); !ok {
		log.Error("the self update %s in the self update desiredstate feedback is not of type int", fieldName)
		return 0, false
	} else {
		return convertedValue, true
	}
}

func (su *selfUpdateOperation) getSelfUpdateFieldString(mapValue map[string]interface{}, mapField string, fieldName string) (string, bool) {
	if value, ok := mapValue[mapField]; !ok {
		log.Error("missing self update %s in the self update desiredstate feedback", fieldName)
		return "", false
	} else if convertedValue, ok := value.(string); !ok {
		log.Error("the self update %s in the self update desiredstate feedback is not of type string", fieldName)
		return "", false
	} else {
		return convertedValue, true
	}
}

func (su *selfUpdateOperation) getSelfUpdateFieldMap(mapValue map[string]interface{}, mapField string, fieldName string) (map[string]interface{}, bool) {
	if value, ok := mapValue[mapField]; !ok {
		log.Error("missing self update %s in the self update desiredstate feedback", fieldName)
		return nil, false
	} else if convertedValue, ok := value.(map[string]interface{}); !ok {
		log.Error("the self update %s in the self update desiredstate feedback is not of type map", fieldName)
		return nil, false
	} else {
		return convertedValue, true
	}
}

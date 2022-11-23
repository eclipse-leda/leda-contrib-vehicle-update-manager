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
	"context"
	"sync"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	topicSelfUpdateDesiredState         = "selfupdate/desiredstate"
	topicSelfUpdateCurrentState         = "selfupdate/currentstate"
	topicSelfUpdateDesiredStateFeedback = "selfupdate/desiredstatefeedback"
)

type selfUpdateManager struct {
	cfg                 *mgrOpts
	applyLock           sync.Mutex
	pahoClient          mqtt.Client
	currentState        *unstructured.Unstructured
	eventsMgr           events.UpdateEventsManager
	selfUpdateOperation *selfUpdateOperation
}

// ApplyResult holds the result of a self update apply
type ApplyResult struct {
	Result         OperationResult
	RebootTimeout  time.Duration
	RebootRequired bool
	Err            error
}

func (suMgr *selfUpdateManager) Apply(ctx context.Context, mf []*unstructured.Unstructured) interface{} {
	suMgr.applyLock.Lock()

	log.Debug("performing self update operation...")

	var applyErr error
	suApplyResult := &ApplyResult{}
	defer func() {
		suMgr.selfUpdateOperation = nil
		if applyErr != nil {
			log.Error(applyErr.Error())
		}
		suMgr.applyLock.Unlock()
	}()

	suMgr.selfUpdateOperation = newSelfUpdateOperation()
	selfUpdateManifest, applyErr := suMgr.unmarshalUnstructured(mf[0])
	if applyErr != nil {
		suApplyResult.Result = SelfUpdateResultError
		suApplyResult.Err = applyErr
		return suApplyResult
	}
	log.Debug("self update manifest: %s", string(selfUpdateManifest))

	if token := suMgr.pahoClient.Subscribe(topicSelfUpdateDesiredStateFeedback, 1, suMgr.selfUpdateOperation.handleSelfUpdateDesiredStateFeedback); !token.WaitTimeout(suMgr.cfg.acknowledgeTimeout) {
		applyErr = log.NewErrorf("cannot subscribe for topic '%s' in '%v' seconds", topicSelfUpdateDesiredStateFeedback, suMgr.cfg.acknowledgeTimeout)
		suApplyResult.Result = SelfUpdateResultError
		suApplyResult.Err = applyErr
		return suApplyResult
	}
	defer suMgr.pahoClient.Unsubscribe(topicSelfUpdateDesiredStateFeedback)

	if token := suMgr.pahoClient.Publish(topicSelfUpdateDesiredState, 1, false, selfUpdateManifest); !token.WaitTimeout(suMgr.cfg.acknowledgeTimeout) {
		applyErr = log.NewErrorf("cannot send the self update manifest to the local broker in '%v' seconds", suMgr.cfg.acknowledgeTimeout)
		suApplyResult.Result = SelfUpdateResultError
		suApplyResult.Err = applyErr
		return suApplyResult
	}

	selfUpdateTimeout := suMgr.convertStringToDuration(suMgr.cfg.timeout, 10*time.Minute)
	select {
	case <-time.After(selfUpdateTimeout):
		applyErr = log.NewErrorf("self update operation is not completed in '%v'", selfUpdateTimeout)
		suApplyResult.Result = SelfUpdateResultTimeout
		suApplyResult.Err = applyErr
		return suApplyResult
	case <-suMgr.selfUpdateOperation.done:
		suApplyResult.Result = suMgr.selfUpdateOperation.result
		if suMgr.selfUpdateOperation.result == SelfUpdateResultError {
			applyErr = suMgr.selfUpdateOperation.err
			suApplyResult.Err = applyErr
		} else if suMgr.selfUpdateOperation.result == SelfUpdateResultInstalled {
			if suMgr.cfg.enableReboot {
				selfUpdateRebootTimeout := suMgr.convertStringToDuration(suMgr.cfg.rebootTimeout, time.Minute)
				suApplyResult.RebootTimeout = selfUpdateRebootTimeout
				suApplyResult.RebootRequired = true
			} else {
				log.Warn("reboot required but automatic rebooting is disabled")
			}
		}
		return suApplyResult
	}
}

func (suMgr *selfUpdateManager) Get(ctx context.Context) []*unstructured.Unstructured {
	return []*unstructured.Unstructured{suMgr.currentState}
}

func (suMgr *selfUpdateManager) Dispose(ctx context.Context) error {
	suMgr.pahoClient.Disconnect(uint(suMgr.cfg.disconnectTimeout))
	return nil
}

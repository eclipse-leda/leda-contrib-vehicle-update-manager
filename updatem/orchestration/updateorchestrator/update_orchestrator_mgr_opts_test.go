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

package updateorchestrator

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
)

func TestMgrOpts(t *testing.T) {
	testCases := map[string]struct {
		opts         []MgrOpt
		expectedOpts *mgrOpts
		expectedErr  error
	}{
		"test_error": {
			opts: []MgrOpt{
				func(mgrOptions *mgrOpts) error {
					return log.NewError("test")
				},
			},
			expectedOpts: nil,
			expectedErr:  log.NewError("test"),
		},
		"test_no_error": {
			opts: []MgrOpt{
				WithConnectionBroker("tcp://localhost:1883"),
				WithConnectionKeepAlive(10000),
				WithConnectionConnectTimeout(30000),
				WithConnectionDisconnectTimeout(200),
				WithConnectionClientUsername("user"),
				WithConnectionClientPassword("pass"),
				WithConnectionAcknowledgeTimeout(20000),
				WithConnectionSubscribeTimeout(20000),
				WithConnectionUnsubscribeTimeout(20000),
			},
			expectedOpts: &mgrOpts{
				broker:             "tcp://localhost:1883",
				keepAlive:          10000,
				disconnectTimeout:  200,
				clientUsername:     "user",
				clientPassword:     "pass",
				acknowledgeTimeout: 20000,
				connectTimeout:     30000,
				subscribeTimeout:   20000,
				unsubscribeTimeout: 20000,
			},
			expectedErr: nil,
		},
	}
	for testCaseName, testCase := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			actualOpts := &mgrOpts{}
			err := applyOptsMgr(actualOpts, testCase.opts...)
			testutil.AssertError(t, testCase.expectedErr, err)
		})
	}
}

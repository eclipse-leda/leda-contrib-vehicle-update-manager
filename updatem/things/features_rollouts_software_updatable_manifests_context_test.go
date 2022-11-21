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
	"context"
	"testing"

	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/util"
)

func TestSetSUInstallContext(t *testing.T) {
	testOpStatus := &datatypes.OperationStatus{
		CorrelationID: "some-correlation-id",
		Status:        datatypes.FinishedSuccess,
		SoftwareModule: &datatypes.SoftwareModuleID{
			Name:    "some-name",
			Version: "1.0.0",
		},
	}
	testCases := map[string]struct {
		ctx           context.Context
		opStatus      *datatypes.OperationStatus
		expectedValue *suOperationContextValue
	}{
		"test_ctx_nil": {
			ctx:           nil,
			opStatus:      testOpStatus,
			expectedValue: nil,
		},
		"test_ctx_value": {
			ctx:      context.Background(),
			opStatus: testOpStatus,
			expectedValue: &suOperationContextValue{
				correlationID:         testOpStatus.CorrelationID,
				softwareModuleName:    testOpStatus.SoftwareModule.Name,
				softwareModuleVersion: testOpStatus.SoftwareModule.Version,
			},
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			t.Log(tcName)
			actualCtx := setSUInstallContext(tc.ctx, tc.opStatus)
			if tc.expectedValue == nil {
				testutil.AssertNil(t, util.GetValue(actualCtx, contextKeyOperationStatus))
			} else {
				testutil.AssertEqual(t, tc.expectedValue, util.GetValue(actualCtx, contextKeyOperationStatus))
			}
		})
	}
}

func TestValidateSUInstallContext(t *testing.T) {
	testOpStatus := &datatypes.OperationStatus{
		CorrelationID: "some-correlation-id",
		Status:        datatypes.FinishedSuccess,
		SoftwareModule: &datatypes.SoftwareModuleID{
			Name:    "some-name",
			Version: "1.0.0",
		},
	}
	testCases := map[string]struct {
		ctx            context.Context
		opStatus       *datatypes.OperationStatus
		expectedResult bool
	}{
		"test_ctx_nil": {
			ctx:            nil,
			opStatus:       testOpStatus,
			expectedResult: false,
		},
		"test_ctx_empty": {
			ctx:            context.Background(),
			opStatus:       testOpStatus,
			expectedResult: false,
		},
		"test_ctx_wrong_value_type": {
			ctx:            context.WithValue(context.Background(), contextKeyOperationStatus, "wrong-value"),
			opStatus:       testOpStatus,
			expectedResult: false,
		},
		"test_ctx_correct": {
			ctx:            setSUInstallContext(context.Background(), testOpStatus),
			opStatus:       testOpStatus,
			expectedResult: true,
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			t.Log(tcName)
			res := validateSUInstallContext(tc.ctx, tc.opStatus)
			testutil.AssertEqual(t, tc.expectedResult, res)
		})
	}
}

func TestGetSUInstallContext(t *testing.T) {
	testOpStatus := &datatypes.OperationStatus{
		CorrelationID: "some-correlation-id",
		SoftwareModule: &datatypes.SoftwareModuleID{
			Name:    "some-name",
			Version: "1.0.0",
		},
	}
	testCases := map[string]struct {
		ctx              context.Context
		expectedOpStatus *datatypes.OperationStatus
	}{
		"test_ctx_nil": {
			ctx:              nil,
			expectedOpStatus: nil,
		},
		"test_ctx_empty": {
			ctx:              context.Background(),
			expectedOpStatus: nil,
		},
		"test_ctx_wrong_value_type": {
			ctx:              context.WithValue(context.Background(), contextKeyOperationStatus, "wrong-value"),
			expectedOpStatus: nil,
		},
		"test_ctx_correct": {
			ctx:              setSUInstallContext(context.Background(), testOpStatus),
			expectedOpStatus: testOpStatus,
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			t.Log(tcName)
			res := getSUInstallContext(tc.ctx)
			if tc.expectedOpStatus == nil {
				testutil.AssertNil(t, res)
			} else {
				testutil.AssertEqual(t, tc.expectedOpStatus, res)
			}
		})
	}
}

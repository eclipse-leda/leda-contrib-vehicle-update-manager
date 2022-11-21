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

package orchestration

import (
	"context"
	"testing"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSetOrchestrationMgrApplyContext(t *testing.T) {
	testMf := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": "test-pod-name2",
				},
			},
		},
	}
	testCases := map[string]struct {
		ctx           context.Context
		manifest      []*unstructured.Unstructured
		expectedValue []*unstructured.Unstructured
	}{
		"test_ctx_nil": {
			ctx:           nil,
			manifest:      testMf,
			expectedValue: nil,
		},
		"test_ctx_value": {
			ctx:           context.Background(),
			manifest:      testMf,
			expectedValue: testMf,
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			t.Log(tcName)
			actualCtx := SetUpdateMgrApplyContext(tc.ctx, tc.manifest)
			if tc.expectedValue == nil {
				testutil.AssertNil(t, util.GetValue(actualCtx, contextKeyManifestInfo))
			} else {
				testutil.AssertEqual(t, tc.expectedValue, util.GetValue(actualCtx, contextKeyManifestInfo))
			}
		})
	}
}

func TestGetOrchestrationMgrApplyContext(t *testing.T) {

	testMf := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       "Pod",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name": "test-pod-name2",
				},
			},
		},
	}
	testCases := map[string]struct {
		ctx        context.Context
		expectedMf []*unstructured.Unstructured
	}{
		"test_ctx_nil": {
			ctx:        nil,
			expectedMf: nil,
		},
		"test_ctx_empty": {
			ctx:        context.Background(),
			expectedMf: nil,
		},
		"test_ctx_wrong_value_type": {
			ctx:        context.WithValue(context.Background(), contextKeyManifestInfo, "wrong-value"),
			expectedMf: nil,
		},
		"test_ctx_correct": {
			ctx:        SetUpdateMgrApplyContext(context.Background(), testMf),
			expectedMf: testMf,
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			t.Log(tcName)
			res := GetUpdateMgrApplyContext(tc.ctx)
			if tc.expectedMf == nil {
				testutil.AssertNil(t, res)
			} else {
				testutil.AssertEqual(t, tc.expectedMf, res)
			}
		})
	}
}

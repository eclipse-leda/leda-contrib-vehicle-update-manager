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
	"context"
	"testing"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
)

func TestContextGetValueOfType(t *testing.T) {
	type testKeyType struct{}
	type testValueType struct{}
	var (
		testKey   = &testKeyType{}
		testValue = &testValueType{}
	)

	testCases := map[string]struct {
		ctx           context.Context
		key           *testKeyType
		expectedValue interface{}
	}{
		"test_nil_ctx": {
			ctx:           nil,
			key:           testKey,
			expectedValue: nil,
		},
		"test_ctx_empty": {
			ctx:           context.Background(),
			key:           testKey,
			expectedValue: nil,
		},
		"test_ctx_value_set": {
			ctx:           context.WithValue(context.Background(), testKey, testValue),
			key:           testKey,
			expectedValue: testValue,
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			t.Log(tcName)
			actualValue := GetValue(tc.ctx, tc.key)
			testutil.AssertEqual(t, tc.expectedValue, actualValue)
		})
	}
}

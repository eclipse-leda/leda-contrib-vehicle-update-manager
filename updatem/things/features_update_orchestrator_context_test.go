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

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
)

func TestApplyCorrelationIDToContextSuccessfully(t *testing.T) {
	testKey := "testKey"
	context := setApplyCorrelationIDContext(context.Background(), testKey)

	testutil.AssertEqual(t, testKey, getApplyCorrelationIDContext(context))
}

func TestSetApplyCorrelationIDToContextWithNilContext(t *testing.T) {
	context := setApplyCorrelationIDContext(nil, "")

	testutil.AssertNil(t, context)
}

func TestGetApplyCorrelationIDToContextWithoutCorrelationID(t *testing.T) {
	testutil.AssertEqual(t, "", getApplyCorrelationIDContext(nil))
}

func TestGetApplyCorrelationIDToContextWithEmptyToString(t *testing.T) {
	context := setApplyCorrelationIDContext(context.Background(), "")

	testutil.AssertEqual(t, "", getApplyCorrelationIDContext(context))
}

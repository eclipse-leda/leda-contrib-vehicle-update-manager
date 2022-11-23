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

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/util"
)

var (
	contextKeyApplyCorrelationID = &applyCorrelationIDContextKey{}
)

type applyCorrelationIDContextKey struct{}

func setApplyCorrelationIDContext(ctx context.Context, correlationID string) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, contextKeyApplyCorrelationID, correlationID)
}

func getApplyCorrelationIDContext(ctx context.Context) string {
	ctxApplyCorrelationID := util.GetValue(ctx, contextKeyApplyCorrelationID)
	if ctxApplyCorrelationID == nil {
		return ""
	}
	applyCorrelationID, ok := ctxApplyCorrelationID.(string)
	if !ok {
		return ""
	}
	return applyCorrelationID
}

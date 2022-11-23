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
)

// GetValue retrieves a specific value from the context based on the provided key
func GetValue(ctx context.Context, key interface{}) interface{} {
	if ctx == nil {
		return nil
	}
	ctxValue := ctx.Value(key)
	if ctxValue == nil {
		return nil
	}
	return ctxValue
}

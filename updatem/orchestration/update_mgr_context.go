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

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	contextKeyManifestInfo = &orchestrationCtxKey{}
)

type orchestrationCtxKey struct{}

// SetUpdateMgrApplyContext ensures the context used throughout a running orchestration
func SetUpdateMgrApplyContext(ctx context.Context, mf []*unstructured.Unstructured) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, contextKeyManifestInfo, mf)
}

// GetUpdateMgrApplyContext retrieves the values related to the context used throughout a running orchestration
func GetUpdateMgrApplyContext(ctx context.Context) []*unstructured.Unstructured {
	ctxMfInfo := util.GetValue(ctx, contextKeyManifestInfo)
	if ctxMfInfo == nil {
		return nil
	}
	mfInfo, ok := ctxMfInfo.([]*unstructured.Unstructured)
	if !ok {
		return nil
	}
	return mfInfo
}

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

	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/util"
)

var (
	contextKeyOperationStatus = &suOperationContextKey{}
)

type suOperationContextKey struct{}

type suOperationContextValue struct {
	correlationID         string
	softwareModuleName    string
	softwareModuleVersion string
}

func setSUInstallContext(ctx context.Context, opStatus *datatypes.OperationStatus) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, contextKeyOperationStatus, &suOperationContextValue{
		correlationID:         opStatus.CorrelationID,
		softwareModuleName:    opStatus.SoftwareModule.Name,
		softwareModuleVersion: opStatus.SoftwareModule.Version,
	})
}

func validateSUInstallContext(ctx context.Context, expectedOpStatus *datatypes.OperationStatus) bool {
	ctxOpStatus := util.GetValue(ctx, contextKeyOperationStatus)
	if ctxOpStatus == nil {
		return false
	}
	opStatus, ok := ctxOpStatus.(*suOperationContextValue)
	if !ok {
		return false
	}
	return opStatus.correlationID == expectedOpStatus.CorrelationID
}

func getSUInstallContext(ctx context.Context) *datatypes.OperationStatus {
	ctxOpStatus := util.GetValue(ctx, contextKeyOperationStatus)
	if ctxOpStatus == nil {
		return nil
	}
	opStatus, ok := ctxOpStatus.(*suOperationContextValue)
	if !ok {
		return nil
	}
	return &datatypes.OperationStatus{
		CorrelationID: opStatus.correlationID,
		SoftwareModule: &datatypes.SoftwareModuleID{
			Name:    opStatus.softwareModuleName,
			Version: opStatus.softwareModuleVersion,
		},
	}
}

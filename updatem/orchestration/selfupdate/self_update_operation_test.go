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
	"testing"

	mocks "github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil/mocks/mqtt"
	"github.com/golang/mock/gomock"
)

func TestMalformedSelfUpdateDesiredStateFeedback(t *testing.T) {
	var testData = []struct {
		err                  string
		desiredStateFeedback string
	}{
		{
			"no_content",
			"",
		},
		{
			"incorrect_yaml_format",
			`
apiVersion: sdv.eclipse.org/v1
kind SelfUpdateBundle
metadata:
	name: self-update-bundle-example
spec:
bundleDownloadUrl: baseURL
bundleName: swdv-arm64-build42
bundleTarget: base
bundleVersion: v1beta3
state:
  message: "Self update bundle installed"
  name: installed
`,
		},
		{
			"missing_metadata",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle installed"
 name: installed
`,
		},
		{
			"missing_state",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
`,
		},
		{
			"missing_state_name",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle installed"
`,
		},
		{
			"invalid_state_name_type",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle installed"
 name: 1000
`,
		},
		{
			"metadata_name_not_string",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: 123
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
 progress: 50
`,
		},
		{
			"state_not_of_type_map",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
`,
		},
	}
	controller := gomock.NewController(t)
	for _, testValues := range testData {
		t.Run(testValues.err, func(t *testing.T) {
			selfUpdateOperation := newSelfUpdateOperation()
			mockClient := mocks.NewMockClient(controller)
			selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, testValues.desiredStateFeedback))
			assertSelfUpdateOperationResult(t, OperationResult(SelfUpdateNoResult), false, selfUpdateOperation)
		})
	}
}

func TestSelfUpdateStateFailed(t *testing.T) {
	var testData = []struct {
		err                  string
		desiredResult        OperationResult
		hasError             bool
		desiredStateFeedback string
	}{
		{
			"download_failed",
			SelfUpdateResultError,
			true,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Cannot download self update bundle"
 name: failed
 techCode: 1001
`,
		},
		{
			"invalid_bundle",
			SelfUpdateResultError,
			true,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Invalid self update bundle"
 name: failed
 techCode: 2001
`,
		},
		{
			"installation_failed",
			SelfUpdateResultError,
			true,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Cannot install self update bundle"
 name: failed
 techCode: 3001
`,
		},
		{
			"selfupdate_rejected",
			SelfUpdateResultRejected,
			false,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Cannot update bundle with the same version"
 name: failed
 techCode: 4001
`,
		},
		{
			"unknown_error",
			SelfUpdateResultError,
			true,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update unknown error"
 name: failed
 techCode: 5001
`,
		},
	}
	controller := gomock.NewController(t)
	for _, testValues := range testData {
		t.Run(testValues.err, func(t *testing.T) {
			selfUpdateOperation := newSelfUpdateOperation()
			mockClient := mocks.NewMockClient(controller)
			selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, testValues.desiredStateFeedback))
			assertSelfUpdateOperationResult(t, OperationResult(testValues.desiredResult), testValues.hasError, selfUpdateOperation)
		})
	}
}

func TestSelfUpdateDesiredState(t *testing.T) {
	var testData = []struct {
		err                  string
		desiredResult        OperationResult
		hasError             bool
		desiredStateFeedback string
	}{
		{
			"state_idle",
			SelfUpdateNoResult,
			false,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle idle"
 name: idle
`,
		},
		{
			"state_failed",
			SelfUpdateResultError,
			true,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle failed"
 name: failed
 techCode: 5001
`,
		},
		{
			"state_installed",
			SelfUpdateResultInstalled,
			false,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle installed"
 name: installed
`,
		},
		{
			"state_installing",
			SelfUpdateNoResult,
			false,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle installing"
 name: installing
 progress: 50
`,
		},
		{
			"state_downloading",
			SelfUpdateNoResult,
			false,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
 progress: 50
`,
		},
		{
			"state_uninitialized",
			SelfUpdateResultError,
			true,
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "SUA is not configured yet"
 name: uninitialized
`,
		},
	}
	controller := gomock.NewController(t)
	for _, testValues := range testData {
		t.Run(testValues.err, func(t *testing.T) {
			selfUpdateOperation := newSelfUpdateOperation()
			mockClient := mocks.NewMockClient(controller)
			selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, testValues.desiredStateFeedback))
			assertSelfUpdateOperationResult(t, OperationResult(testValues.desiredResult), testValues.hasError, selfUpdateOperation)
		})
	}
}
func TestSelfUpdateOperationStateInstalled(t *testing.T) {
	selfUpdateDesiredStateFeedbackDownloading := `
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
 progress: 50
`
	selfUpdateDesiredStateFeedbackInstalling := `
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle installing"
 name: installing
 progress: 50
`
	selfUpdateDesiredStateFeedbackInstalled := `
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle installed"
 name: installed
`
	controller := gomock.NewController(t)

	selfUpdateOperation := newSelfUpdateOperation()
	mockClient := mocks.NewMockClient(controller)

	selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, selfUpdateDesiredStateFeedbackDownloading))
	assertSelfUpdateOperationResult(t, SelfUpdateNoResult, false, selfUpdateOperation)

	selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, selfUpdateDesiredStateFeedbackInstalling))
	assertSelfUpdateOperationResult(t, SelfUpdateNoResult, false, selfUpdateOperation)

	selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, selfUpdateDesiredStateFeedbackInstalled))
	assertSelfUpdateOperationResult(t, SelfUpdateResultInstalled, false, selfUpdateOperation)
}

func TestInvalidSelfUpdateStateTechCode(t *testing.T) {
	var testData = []struct {
		err                  string
		desiredStateFeedback string
	}{
		{
			"missing_tech_code",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update failed"
 name: failed
`,
		},
		{
			"invalid_type",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update failed"
 name: failed
 techCode: "1001"
`,
		},
		{
			"unsupported_tech_code",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update failed"
 name: failed
 techCode: 7001
`,
		},
		{
			"out_of_range_progress",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
 progress: 110
`,
		},
	}
	controller := gomock.NewController(t)
	for _, testValues := range testData {
		t.Run(testValues.err, func(t *testing.T) {
			selfUpdateOperation := newSelfUpdateOperation()
			mockClient := mocks.NewMockClient(controller)
			selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, testValues.desiredStateFeedback))
			assertSelfUpdateOperationResult(t, SelfUpdateNoResult, false, selfUpdateOperation)
		})
	}
}

func TestInvalidSelfUpdateProgress(t *testing.T) {
	var testData = []struct {
		err                  string
		desiredStateFeedback string
	}{
		{
			"missing_progress",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
`,
		},
		{
			"invalid_type",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
 progress: "50"
`,
		},
		{
			"negative_progress",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
 progress: -1
`,
		},
		{
			"out_of_range_progress",
			`
apiVersion: sdv.eclipse.org/v1
kind: SelfUpdateBundle
metadata:
 name: self-update-bundle-example
spec:
 bundleDownloadUrl: baseURL
 bundleName: swdv-arm64-build42
 bundleTarget: base
 bundleVersion: v1beta3
state:
 message: "Self update bundle downloading"
 name: downloading
 progress: 110
`,
		},
	}
	controller := gomock.NewController(t)
	for _, testValues := range testData {
		t.Run(testValues.err, func(t *testing.T) {
			selfUpdateOperation := newSelfUpdateOperation()
			mockClient := mocks.NewMockClient(controller)
			selfUpdateOperation.handleSelfUpdateDesiredStateFeedback(mockClient, setupMockMessage(controller, testValues.desiredStateFeedback))
			assertSelfUpdateOperationResult(t, SelfUpdateNoResult, false, selfUpdateOperation)
		})
	}
}

func assertSelfUpdateOperationResult(t *testing.T, expectedResult OperationResult, hasError bool, su *selfUpdateOperation) {
	if su.result != expectedResult {
		t.Fail()
	}
	if hasError && (su.err == nil) {
		t.Fail()
	}
	if !hasError && (su.err != nil) {
		t.Fail()
	}
}

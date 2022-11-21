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

type manifestStatus string

const (
	manifestStatusStarted          manifestStatus = "STARTED"
	manifestStatusRunning          manifestStatus = "RUNNING"
	manifestStatusModifiedLocally  manifestStatus = "MODIFIED_LOCALLY"
	manifestStatusFinishedSuccess  manifestStatus = "FINISHED_SUCCESS"
	manifestStatusFinishedError    manifestStatus = "FINISHED_ERROR"
	manifestStatusFinishedRejected manifestStatus = "FINISHED_REJECTED"
)

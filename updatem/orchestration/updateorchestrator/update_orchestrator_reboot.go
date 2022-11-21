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

package updateorchestrator

import (
	"os"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

const (
	procSysRqFile        = "/proc/sys/kernel/sysrq"
	procSysRqTriggerFile = "/proc/sysrq-trigger"
)

// RebootManager defines an interface for restarting the host system
type RebootManager interface {
	Reboot(time.Duration) error
}

type rebootManager struct{}

func (r *rebootManager) Reboot(timeout time.Duration) error {
	log.Debug("the system is about to reboot after successful update operation in '%s'", timeout)
	<-time.After(timeout)
	if err := os.WriteFile(procSysRqFile, []byte("1"), 0644); err != nil {
		return log.NewErrorf("cannot reboot after successful update operation. cannot send signal to %v", procSysRqFile)
	}
	if err := os.WriteFile(procSysRqTriggerFile, []byte("b"), 0200); err != nil {
		return log.NewErrorf("cannot reboot after successful update operation. cannot send signal to %v", procSysRqTriggerFile)
	}
	return nil
}

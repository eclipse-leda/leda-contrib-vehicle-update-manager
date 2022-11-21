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
	"time"
)

// MgrOpt defines the creation configuration options for a self update manager implementation
type MgrOpt func(mgrOptions *mgrOpts) error

type mgrOpts struct {
	enableReboot       bool
	rebootTimeout      string
	timeout            string
	broker             string
	keepAlive          time.Duration
	disconnectTimeout  time.Duration
	clientUsername     string
	clientPassword     string
	connectTimeout     time.Duration
	acknowledgeTimeout time.Duration
	subscribeTimeout   time.Duration
	unsubscribeTimeout time.Duration
}

func applyOptsMgr(mgrOpts *mgrOpts, opts ...MgrOpt) error {
	for _, o := range opts {
		if err := o(mgrOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithEnableReboot configures the required reboot parameter to the self update manager runtime configuration
func WithEnableReboot(reboot bool) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.enableReboot = reboot
		return nil
	}
}

// WithTimeout configures the timeout parameter to the self update manager runtime configuration
func WithTimeout(timeout string) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.timeout = timeout
		return nil
	}
}

// WithRebootTimeout configures the reboot timeout parameter to the self update manager runtime configuration
func WithRebootTimeout(rebootTimeout string) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.rebootTimeout = rebootTimeout
		return nil
	}
}

// WithConnectionBroker configures the broker, where the connection will be established
func WithConnectionBroker(broker string) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.broker = broker
		return nil
	}
}

// WithConnectionKeepAlive configures the time between between each check for the connection presence
func WithConnectionKeepAlive(keepAlive time.Duration) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.keepAlive = keepAlive
		return nil
	}
}

// WithConnectionDisconnectTimeout configures the duration of inactivity before disconnecting from the broker
func WithConnectionDisconnectTimeout(disconnectTimeout time.Duration) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.disconnectTimeout = disconnectTimeout
		return nil
	}
}

// WithConnectionClientUsername configures the client username used when establishing connection to the broker
func WithConnectionClientUsername(username string) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.clientUsername = username
		return nil
	}
}

// WithConnectionClientPassword configures the client password used when establishing connection to the broker
func WithConnectionClientPassword(password string) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.clientPassword = password
		return nil
	}
}

// WithConnectionConnectTimeout configures the timeout before terminating the connect attempt
func WithConnectionConnectTimeout(connectTimeout time.Duration) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.connectTimeout = connectTimeout
		return nil
	}
}

// WithConnectionAcknowledgeTimeout configures the timeout for the acknowledge receival
func WithConnectionAcknowledgeTimeout(acknowledgeTimeout time.Duration) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.acknowledgeTimeout = acknowledgeTimeout
		return nil
	}
}

// WithConnectionSubscribeTimeout configures the timeout before terminating the subscribe attempt
func WithConnectionSubscribeTimeout(subscribeTimeout time.Duration) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.subscribeTimeout = subscribeTimeout
		return nil
	}
}

// WithConnectionUnsubscribeTimeout configures the timeout before terminating the unsubscribe attempt
func WithConnectionUnsubscribeTimeout(unsubscribeTimeout time.Duration) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.unsubscribeTimeout = unsubscribeTimeout
		return nil
	}
}

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

package k8s

// MgrOpt defines the creation configuration options for a k8s API-compatible orchestration implementation
type MgrOpt func(mgrOptions *mgrOpts) error

type mgrOpts struct {
	kubeconfig string
}

func applyOptsMgr(mgrOpts *mgrOpts, opts ...MgrOpt) error {
	for _, o := range opts {
		if err := o(mgrOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithKubeConfig configures the absolute path to the k8s runtime configuration
func WithKubeConfig(kubeconfig string) MgrOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.kubeconfig = kubeconfig
		return nil
	}
}

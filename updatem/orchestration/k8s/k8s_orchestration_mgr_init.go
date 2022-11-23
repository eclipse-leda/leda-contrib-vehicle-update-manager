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

import (
	"context"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// NewK8sUpdateManager instantiates a new k8s update manager
func NewK8sUpdateManager(opts []MgrOpt, registryCtx *registry.ServiceRegistryContext) (orchestration.UpdateManager, error) {
	var (
		cfg = &mgrOpts{}
	)
	applyOptsMgr(cfg, opts...)

	eventsManagerService, errEvtsMgr := registryCtx.Get(registry.EventsManagerService)
	if errEvtsMgr != nil {
		return nil, errEvtsMgr
	}

	config, err := clientcmd.BuildConfigFromFlags("", cfg.kubeconfig)

	if err != nil {
		log.ErrorErr(err, "error parsing the provided kubeconfig")
		return nil, err
	}
	k8sClient, k8sInitErr := dynamic.NewForConfig(config)
	if k8sInitErr != nil {
		log.ErrorErr(err, "error connecting to the provided k8s instance")
		return nil, k8sInitErr
	}
	log.Debug("successfully connected via the provided kubeconfig %s", cfg.kubeconfig)

	k8sDiscoveryClient, dcErr := discovery.NewDiscoveryClientForConfig(config)
	if dcErr != nil {
		log.ErrorErr(dcErr, "error connecting k8s discovery client to the provided k8s instance")
		return nil, dcErr
	}
	k8sRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(k8sDiscoveryClient))
	log.Debug("successfully initialized k8s REST mapper instance")

	k8sOrchMgr := &k8sUpdateManager{
		cfg:                      cfg,
		k8sClient:                k8sClient,
		k8sRESTMapper:            k8sRESTMapper,
		eventsMgr:                eventsManagerService.(events.UpdateEventsManager),
		flagPublishResourceEvent: true,
	}

	k8sOrchMgr.loopWatchResources(context.Background())

	return k8sOrchMgr, nil
}

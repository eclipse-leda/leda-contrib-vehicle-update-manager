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
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// UpdateManagerK8sServiceID is the ID os the locally registered k8s update manager implementation
const UpdateManagerK8sServiceID = "orchestration.mgr"

type k8sUpdateManager struct {
	cfg                      *mgrOpts
	k8sClient                dynamic.Interface
	k8sRESTMapper            meta.RESTMapper
	eventsMgr                events.UpdateEventsManager
	applyLock                sync.Mutex
	flagPublishResourceEvent bool
}

var (
	coreV1PodGVK = schema.GroupVersionKind{
		Group:   "", // core API does not have a group
		Version: "v1",
		Kind:    "Pod",
	}
)

func (updMgr *k8sUpdateManager) Apply(ctx context.Context, mf []*unstructured.Unstructured) interface{} {
	updMgr.applyLock.Lock()
	updMgr.flagPublishResourceEvent = false

	log.Debug("processing apply manifest - start")

	defer func() {
		updMgr.flagPublishResourceEvent = true
		updMgr.applyLock.Unlock()
	}()

	cmd, err := newKubectlApply(&updMgr.cfg.kubeconfig)
	if err != nil {
		log.Error("error while creating apply manifest command ", err)
		return err
	}

	err = cmd.apply(mf)
	if err != nil {
		log.Error("error while applying manifest ", err)
		return err
	}

	log.Debug("finished applying manifest")
	return nil
}

func (updMgr *k8sUpdateManager) Dispose(ctx context.Context) error {
	return nil
}

func (updMgr *k8sUpdateManager) Get(ctx context.Context) []*unstructured.Unstructured {
	updMgr.applyLock.Lock()

	log.Debug("list existing k8s pods & nodes ")
	defer updMgr.applyLock.Unlock()

	// List pods
	podsUnstructured, err := updMgr.listAllResources(ctx, coreV1PodGVK)
	if err != nil {
		log.Error("cannot list the existing pods ", err)
		return nil
	}

	// List nodes
	nodesUnstructured, err := updMgr.listAllResources(ctx, coreV1NodeGVK)
	if err != nil {
		log.Error("cannot list the existing nodes ", err)
		return nil
	}

	result := []*unstructured.Unstructured{}
	result = append(result, podsUnstructured...)
	result = append(result, nodesUnstructured...)
	return result
}

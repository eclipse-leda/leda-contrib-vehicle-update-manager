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
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

var coreV1NodeGVK = schema.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "Node",
}

func (updMgr *k8sUpdateManager) publishResourceEvent(ctx context.Context, eventAction events.EventAction, eventSource unstructured.Unstructured, err error) {
	e := &events.Event{
		Type:    events.EventTypeResources,
		Action:  eventAction,
		Source:  eventSource,
		Time:    time.Now().UTC().Unix(),
		Context: ctx,
		Error:   err,
	}

	if updMgr.flagPublishResourceEvent {
		//log.Trace("publishing resource event [%+v]", e)
		if pubErr := updMgr.eventsMgr.Publish(ctx, e); pubErr != nil {
			log.ErrorErr(pubErr, "failed to publish resource event [%+v]", e)
		}
		return
	}
	//log.Trace("publishing resource event skipped, ongoing apply")
}

// k8s apimachinery wrapper -------------------------------------

func (updMgr *k8sUpdateManager) getResourceByGVK(gvk schema.GroupVersionKind, namespace string) (dynamic.ResourceInterface, error) {
	mapping, mappingErr := updMgr.k8sRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if mappingErr != nil {
		return nil, mappingErr
	}
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		return updMgr.k8sClient.Resource(mapping.Resource).Namespace(namespace), nil
	}
	log.Debug("the requested GroupVersionKind [%s] resource is not namespaced - will not use the provided namespace [%s]", gvk.String(), namespace)
	return updMgr.k8sClient.Resource(mapping.Resource), nil
}

func (updMgr *k8sUpdateManager) listAllResources(ctx context.Context, gvk schema.GroupVersionKind) ([]*unstructured.Unstructured, error) {
	result := []*unstructured.Unstructured{}

	resourceByNs, err := updMgr.getResourceByGVK(gvk, "")
	if err != nil {
		return nil, err
	}
	if resourceByNs != nil {
		resourcesByNs, err := resourceByNs.List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, resourceUnstructured := range resourcesByNs.Items {
			result = append(result, resourceUnstructured.DeepCopy())
		}
	}
	return result, nil
}

func (updMgr *k8sUpdateManager) loopWatchResources(ctx context.Context) {
	updMgr.loopWatchResource(ctx, "pods.v1.")
	updMgr.loopWatchResource(ctx, "nodes.v1.")
}

func (updMgr *k8sUpdateManager) loopWatchResource(ctx context.Context, resource string) {

	log.Debug("Start watching %s ...", resource)

	resyncPeriod := 0 * time.Minute
	di := dynamicinformer.NewDynamicSharedInformerFactory(updMgr.k8sClient, resyncPeriod)
	// Retrieve a "GroupVersionResource" type that needed when generating informer from dynamic factory
	gvr, _ := schema.ParseResourceArg(resource) //pods.v1. //nodes.v1.
	// Create informer
	i := di.ForResource(*gvr)

	i.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				log.Debug("Received add event! %s - %s ", u.GetNamespace(), u.GetName())
				updMgr.publishResourceEvent(ctx, events.EventActionResourcesAdded, *u, nil)
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				//uOld := oldObj.(*unstructured.Unstructured)
				uNew := newObj.(*unstructured.Unstructured)
				//log.Debug("Received update Old event! %s - %s ", uOld.GetNamespace(), uOld.GetName())
				//log.Debug("%s\n",oldObj)
				//log.Debug("Received update New event! %s - %s ", uNew.GetNamespace(), uNew.GetName())
				//log.Debug("%s\n",uNew)
				updMgr.publishResourceEvent(ctx, events.EventActionResourcesUpdated, *uNew, nil)
			},
			DeleteFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				log.Debug("Received delete event! %s - %s ", u.GetNamespace(), u.GetName())
				updMgr.publishResourceEvent(ctx, events.EventActionResourcesDeleted, *u, nil)
			},
		},
	)

	di.Start(wait.NeverStop)
}

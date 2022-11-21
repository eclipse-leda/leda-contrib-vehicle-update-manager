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
	"sync"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/events"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-kanto/container-management/rollouts/api/features"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
)

// Constants the define the respective Ditto feature based on the SU v2 Vorto model
const (
	SoftwareUpdatableManifestsFeatureID       = "SoftwareUpdatable:manifest"
	updateSoftwareUpdatableManifestsAgentType = "manifest"

	softwareUpdatableDefinitionNamespace = "org.eclipse.hawkbit.swupdatable"
	softwareUpdatableDefinitionName      = "SoftwareUpdatable"
	softwareUpdatableDefinitionVersion   = "2.0.0"
	softwareUpdatablePropertyNameStatus  = "status"

	softwareUpdatablePropertyLastOperation       = softwareUpdatablePropertyNameStatus + "/lastOperation"
	softwareUpdatablePropertyLastFailedOperation = softwareUpdatablePropertyNameStatus + "/lastFailedOperation"
	softwareUpdatableOperationInstall            = "install"
)

type softwareUpdatableManifests struct {
	rootThing             model.Thing
	status                *features.SoftwareUpdatableStatus
	eventsMgr             events.UpdateEventsManager
	orchMgr               orchestration.UpdateManager
	processOperationsLock sync.Mutex
	statusUpdatesLock     sync.RWMutex
	cancelEventsHandler   context.CancelFunc
}

func (suMf *softwareUpdatableManifests) createFeature() model.Feature {
	feature := client.NewFeature(SoftwareUpdatableManifestsFeatureID,
		client.WithFeatureDefinition(client.NewDefinitionID(softwareUpdatableDefinitionNamespace, softwareUpdatableDefinitionName, softwareUpdatableDefinitionVersion)),
		client.WithFeatureProperty(softwareUpdatablePropertyNameStatus, suMf.status),
		client.WithFeatureOperationsHandler(suMf.operationsHandler))
	return feature
}

func newSoftwareUpdatableManifests(rootThing model.Thing, eventsMgr events.UpdateEventsManager, orchMgr orchestration.UpdateManager) managedFeature {
	supStatus := &features.SoftwareUpdatableStatus{
		SoftwareModuleType: updateSoftwareUpdatableManifestsAgentType,
	}
	return &softwareUpdatableManifests{
		rootThing: rootThing,
		status:    supStatus,
		eventsMgr: eventsMgr,
		orchMgr:   orchMgr,
	}
}

func (suMf *softwareUpdatableManifests) register(ctx context.Context) error {
	log.Debug("initializing SoftwareUpdatable:manifest feature")

	if suMf.cancelEventsHandler == nil {
		suMf.handleOrchestrationEvents(ctx)
		log.Debug("subscribed for update manager events")
	}
	return suMf.rootThing.SetFeature(SoftwareUpdatableManifestsFeatureID, suMf.createFeature())
}

func (suMf *softwareUpdatableManifests) dispose() {
	log.Debug("disposing SoftwareUpdatable feature")
	if suMf.cancelEventsHandler != nil {
		log.Debug("unsubscribing from update manager events")
		suMf.cancelEventsHandler()
		suMf.cancelEventsHandler = nil
	}
}

func (suMf *softwareUpdatableManifests) operationsHandler(operationName string, args interface{}) (interface{}, error) {
	log.Debug("manifests operation initiated - [operation = %s]", operationName)
	switch operationName {
	case softwareUpdatableOperationInstall:
		ua, err := convertToUpdateAction(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, suMf.install(ua)
	default:
		err := log.NewErrorf("unsupported operation called [operationId = %s]", operationName)
		log.ErrorErr(err, "unsupported operation")
		return nil, client.NewMessagesSubjectNotFound(err.Error())
	}
}
func (suMf *softwareUpdatableManifests) install(updateAction datatypes.UpdateAction) error {
	log.Debug("will perform installation...")
	if err := validateSoftwareUpdateActionManifests(updateAction); err != nil {
		return client.NewMessagesParameterInvalidError(err.Error())
	}
	go suMf.processUpdateAction(updateAction)
	return nil
}

func (suMf *softwareUpdatableManifests) updateLastOperation(operationStatus *datatypes.OperationStatus) {
	suMf.statusUpdatesLock.Lock()
	defer suMf.statusUpdatesLock.Unlock()
	suMf.status.LastOperation = operationStatus
	err := suMf.rootThing.SetFeatureProperty(SoftwareUpdatableManifestsFeatureID, softwareUpdatablePropertyLastOperation, operationStatus)
	if err != nil {
		log.ErrorErr(err, "error while updating lastOperation property")
	}
}

func (suMf *softwareUpdatableManifests) updateLastFailedOperation(operationStatus *datatypes.OperationStatus) {
	suMf.statusUpdatesLock.Lock()
	defer suMf.statusUpdatesLock.Unlock()
	suMf.status.LastFailedOperation = operationStatus
	err := suMf.rootThing.SetFeatureProperty(SoftwareUpdatableManifestsFeatureID, softwareUpdatablePropertyLastFailedOperation, operationStatus)
	if err != nil {
		log.ErrorErr(err, "error while updating lastFailedOperation property")
	}
}
func (suMf *softwareUpdatableManifests) getLastOperation() *datatypes.OperationStatus {
	suMf.statusUpdatesLock.RLock()
	defer suMf.statusUpdatesLock.RUnlock()
	return suMf.status.LastOperation
}

func (suMf *softwareUpdatableManifests) validateOperationContext(ctx context.Context) bool {
	return validateSUInstallContext(ctx, suMf.getLastOperation())
}

func (suMf *softwareUpdatableManifests) setOperationContext(ctx context.Context) context.Context {
	return setSUInstallContext(ctx, suMf.getLastOperation())
}

func (suMf *softwareUpdatableManifests) processUpdateAction(updateAction datatypes.UpdateAction) {
	suMf.processOperationsLock.Lock()
	defer suMf.processOperationsLock.Unlock()

	suMf.installModule(updateAction.SoftwareModules[0], updateAction.CorrelationID)
}

func (suMf *softwareUpdatableManifests) installModule(softMod *datatypes.SoftwareModuleAction, correlationID string) {
	log.Debug("will perform installation of SoftwareModule [Name.version] = [%s.%s]", softMod.SoftwareModule.Name, softMod.SoftwareModule.Version)

	operationStatus := &datatypes.OperationStatus{
		Status:         datatypes.Started,
		CorrelationID:  correlationID,
		SoftwareModule: softMod.SoftwareModule,
	}

	var (
		installError error
		rejected     bool
		mf           []*unstructured.Unstructured
	)

	defer func() {
		// in case of panic report FinishedError
		if err := recover(); err != nil {
			log.Error("failed to install SoftwareModule [Name.version] = [%s.%s]  %v", softMod.SoftwareModule.Name, softMod.SoftwareModule.Version, err)
			operationStatus.Message = "internal runtime error"
			operationStatus.Status = datatypes.FinishedError
			suMf.updateLastFailedOperation(operationStatus)
			suMf.updateLastOperation(operationStatus)
		}
	}()

	suMf.updateLastOperation(operationStatus)

	// Downloading
	operationStatus.Status = datatypes.Downloading
	suMf.updateLastOperation(operationStatus)

	if mf, rejected, installError = getUpdateManifest(softMod.Artifacts[0]); installError != nil {
		log.ErrorErr(installError, "failed to create update manifest from the provided SoftwareArtifact [FileName] = [%s]", softMod.Artifacts[0].FileName)
		operationStatus.Message = installError.Error()
		if rejected {
			operationStatus.Status = datatypes.FinishedRejected
		} else {
			operationStatus.Status = datatypes.FinishedError
		}
		suMf.updateLastFailedOperation(operationStatus)
		suMf.updateLastOperation(operationStatus)
		return
	}

	// Downloaded
	operationStatus.Status = datatypes.Downloaded
	suMf.updateLastOperation(operationStatus)

	ctx := suMf.setOperationContext(context.Background())
	suMf.orchMgr.Apply(ctx, mf)
}

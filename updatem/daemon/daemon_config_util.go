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

package main

import (
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/k8s"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/selfupdate"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/orchestration/updateorchestrator"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/things"

	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/pflag"
)

const daemonConfigFileFlagID = "cfg-file"

func extractUpdateManagerOptions(daemonConfig *config) interface{} {
	mgrOpts := map[string]interface{}{}
	mgrOpts["k8s"] = extractUpdateManagerK8sOptions(daemonConfig)
	mgrOpts["self_update"] = extractUpdateManagerSelfUpdateOptions(daemonConfig)
	mgrOpts["update_orchestrator"] = extractUpdateOrchestratorOptions(daemonConfig)
	return mgrOpts
}

func extractUpdateManagerK8sOptions(daemonConfig *config) []k8s.MgrOpt {
	mgrOpts := []k8s.MgrOpt{}
	mgrOpts = append(mgrOpts,
		k8s.WithKubeConfig(daemonConfig.Orchestration.K8s.Kubeconfig),
	)
	return mgrOpts
}

func extractUpdateManagerSelfUpdateOptions(daemonConfig *config) []selfupdate.MgrOpt {
	mgrOpts := []selfupdate.MgrOpt{}
	mgrOpts = append(mgrOpts,
		selfupdate.WithEnableReboot(daemonConfig.Orchestration.SelfUpdate.EnableReboot),
		selfupdate.WithRebootTimeout(daemonConfig.Orchestration.SelfUpdate.RebootTimeout),
		selfupdate.WithTimeout(daemonConfig.Orchestration.SelfUpdate.Timeout),
		selfupdate.WithConnectionBroker(daemonConfig.ThingsConfig.ThingsConnectionConfig.BrokerURL),
		selfupdate.WithConnectionKeepAlive(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.KeepAlive)*time.Millisecond),
		selfupdate.WithConnectionAcknowledgeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout)*time.Millisecond),
		selfupdate.WithConnectionDisconnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout)*time.Millisecond),
		selfupdate.WithConnectionClientUsername(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientUsername),
		selfupdate.WithConnectionClientPassword(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientPassword),
		selfupdate.WithConnectionConnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.ConnectTimeout)*time.Millisecond),
		selfupdate.WithConnectionSubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout)*time.Millisecond),
		selfupdate.WithConnectionUnsubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout)*time.Millisecond),
	)
	return mgrOpts
}

func extractUpdateOrchestratorOptions(daemonConfig *config) []updateorchestrator.MgrOpt {
	mgrOpts := []updateorchestrator.MgrOpt{}
	mgrOpts = append(mgrOpts,
		updateorchestrator.WithConnectionBroker(daemonConfig.ThingsConfig.ThingsConnectionConfig.BrokerURL),
		updateorchestrator.WithConnectionKeepAlive(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.KeepAlive)*time.Millisecond),
		updateorchestrator.WithConnectionAcknowledgeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout)*time.Millisecond),
		updateorchestrator.WithConnectionDisconnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout)*time.Millisecond),
		updateorchestrator.WithConnectionClientUsername(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientUsername),
		updateorchestrator.WithConnectionClientPassword(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientPassword),
		updateorchestrator.WithConnectionConnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.ConnectTimeout)*time.Millisecond),
		updateorchestrator.WithConnectionSubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout)*time.Millisecond),
		updateorchestrator.WithConnectionUnsubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout)*time.Millisecond),
	)
	return mgrOpts
}

func extractThingsOptions(daemonConfig *config) []things.UpdateThingsManagerOpt {
	thingsOpts := []things.UpdateThingsManagerOpt{}
	thingsOpts = append(thingsOpts,
		things.WithMetaPath(daemonConfig.ThingsConfig.ThingsMetaPath),
		things.WithFeatures(daemonConfig.ThingsConfig.Features),
		things.WithConnectionBroker(daemonConfig.ThingsConfig.ThingsConnectionConfig.BrokerURL),
		things.WithConnectionKeepAlive(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.KeepAlive)*time.Millisecond),
		things.WithConnectionDisconnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout)*time.Millisecond),
		things.WithConnectionClientUsername(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientUsername),
		things.WithConnectionClientPassword(daemonConfig.ThingsConfig.ThingsConnectionConfig.ClientPassword),
		things.WithConnectionConnectTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.ConnectTimeout)*time.Millisecond),
		things.WithConnectionAcknowledgeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout)*time.Millisecond),
		things.WithConnectionSubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout)*time.Millisecond),
		things.WithConnectionUnsubscribeTimeout(time.Duration(daemonConfig.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout)*time.Millisecond),
	)
	return thingsOpts
}

func initLogger(daemonConfig *config) {
	log.Configure(daemonConfig.Log)
}

func loadLocalConfig(filePath string, localConfig *config) error {

	fi, fierr := os.Stat(filePath)
	if fierr != nil {
		if os.IsNotExist(fierr) {
			return nil
		}
		return fierr
	} else if fi.IsDir() {
		return log.NewErrorf("provided configuration path %s is a directory", filePath)
	} else if fi.Size() == 0 {
		log.Warn("the file %s is empty", filePath)
		return nil
	} else {
		log.Debug("successfully retrieved file %s stats", filePath)
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, localConfig)
	if err != nil {
		return err
	}
	return nil
}

func parseConfigFilePath() string {
	var cfgFilePath string
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
	flagSet.StringVar(&cfgFilePath, daemonConfigFileFlagID, daemonConfigFileDefault, "Specify the configuration file of updatemanagerd")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Info("there are flags set for starting the update manager instance configuration from the command line - file and default configurations will be overridden")
	}
	log.Info("local daemon configuration is set to %s", cfgFilePath)
	return cfgFilePath
}

func dumpConfiguration(configInstance *config) {
	if configInstance == nil {
		return
	}
	// dump log config
	dumpLog(configInstance)

	// dump things client config
	dumpThingsClient(configInstance)

	// dump orchestration config
	dumpOrchestration(configInstance)
}

func dumpLog(configInstance *config) {
	if configInstance.Log != nil {
		if configInstance.Log.LogFile != "" {
			log.Debug("[daemon_cfg][log-file] : %s", configInstance.Log.LogFile)
		}
		log.Debug("[daemon_cfg][log-level] : %v", configInstance.Log.LogLevel)
		log.Debug("[daemon_cfg][log-file-size] : %v", configInstance.Log.LogFileSize)
		log.Debug("[daemon_cfg][log-file-count] : %v", configInstance.Log.LogFileCount)
		log.Debug("[daemon_cfg][log-file-max-age] : %v", configInstance.Log.LogFileMaxAge)
		log.Debug("[daemon_cfg][log-syslog] : %v", configInstance.Log.Syslog)
	}
}

func dumpThingsClient(configInstance *config) {
	if configInstance.ThingsConfig != nil {
		log.Debug("[daemon_cfg][things-home-dir] : %s", configInstance.ThingsConfig.ThingsMetaPath)
		log.Debug("[daemon_cfg][things-features] : %s", configInstance.ThingsConfig.Features)
		if configInstance.ThingsConfig.ThingsConnectionConfig != nil {
			log.Debug("[daemon_cfg][things-conn-broker] : %s", configInstance.ThingsConfig.ThingsConnectionConfig.BrokerURL)
			log.Debug("[daemon_cfg][things-conn-keep-alive] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.KeepAlive)
			log.Debug("[daemon_cfg][things-conn-disconnect-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout)
			log.Debug("[daemon_cfg][things-conn-connect-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.ConnectTimeout)
			log.Debug("[daemon_cfg][things-conn-ack-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout)
			log.Debug("[daemon_cfg][things-conn-sub-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout)
			log.Debug("[daemon_cfg][things-conn-unsub-timeout] : %d", configInstance.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout)
		}
	}
}

func dumpOrchestration(configInstance *config) {
	if configInstance.Orchestration != nil {
		log.Debug("[daemon_cfg][k8s-kubeconfig] : %v", configInstance.Orchestration.K8s.Kubeconfig)
		log.Debug("[daemon_cfg][self-update-enable-reboot] : %v", configInstance.Orchestration.SelfUpdate.EnableReboot)
		log.Debug("[daemon_cfg][self-update-timeout] : %v", configInstance.Orchestration.SelfUpdate.Timeout)
		log.Debug("[daemon_cfg][self-update-reboot-timeout] : %v", configInstance.Orchestration.SelfUpdate.RebootTimeout)
	}
}

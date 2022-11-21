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
	"github.com/spf13/cobra"
)

func setupCommandFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	// init daemon general config flag
	flagSet.String(daemonConfigFileFlagID, "", "Specify the configuration file of updatemanagerd")

	// init log flags
	flagSet.StringVar(&cfg.Log.LogLevel, "log-level", cfg.Log.LogLevel, "Set the daemon's log level - possible values are ERROR, WARN, INFO, DEBUG, TRACE")
	flagSet.StringVar(&cfg.Log.LogFile, "log-file", cfg.Log.LogFile, "Set the daemon's log file")
	flagSet.IntVar(&cfg.Log.LogFileSize, "log-file-size", cfg.Log.LogFileSize, "Set the maximum size in megabytes of the log file before it gets rotated")
	flagSet.IntVar(&cfg.Log.LogFileCount, "log-file-count", cfg.Log.LogFileCount, "Set the maximum number of old log files to retain")
	flagSet.IntVar(&cfg.Log.LogFileMaxAge, "log-file-max-age", cfg.Log.LogFileMaxAge, "Set the maximum number of days to retain old log files based on the timestamp encoded in their filename")
	flagSet.BoolVar(&cfg.Log.Syslog, "log-syslog", cfg.Log.Syslog, "Enable logging in the local syslog (e.g. /dev/log, /var/run/syslog, /var/run/log)")

	// init things client
	flagSet.StringVar(&cfg.ThingsConfig.ThingsMetaPath, "things-home-dir", cfg.ThingsConfig.ThingsMetaPath, "Specify the home directory for the Things Update Manager persistent storage")
	flagSet.StringSliceVar(&cfg.ThingsConfig.Features, "things-features", cfg.ThingsConfig.Features, "Specify the desired Ditto features that will be registered for the Ditto thing")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.BrokerURL, "things-conn-broker", cfg.ThingsConfig.ThingsConnectionConfig.BrokerURL, "Specify the MQTT broker URL to connect to")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.KeepAlive, "things-conn-keep-alive", cfg.ThingsConfig.ThingsConnectionConfig.KeepAlive, "Specify the keep alive duration for the MQTT requests in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout, "things-conn-disconnect-timeout", cfg.ThingsConfig.ThingsConnectionConfig.DisconnectTimeout, "Specify the disconnection timeout for the MQTT connection in milliseconds")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.ClientUsername, "things-conn-client-username", cfg.ThingsConfig.ThingsConnectionConfig.ClientUsername, "Specify the MQTT client username to authenticate with")
	flagSet.StringVar(&cfg.ThingsConfig.ThingsConnectionConfig.ClientPassword, "things-conn-client-password", cfg.ThingsConfig.ThingsConnectionConfig.ClientPassword, "Specify the MQTT client password to authenticate with")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.ConnectTimeout, "things-conn-connect-timeout", cfg.ThingsConfig.ThingsConnectionConfig.ConnectTimeout, "Specify the connect timeout for the MQTT in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout, "things-conn-ack-timeout", cfg.ThingsConfig.ThingsConnectionConfig.AcknowledgeTimeout, "Specify the acknowledgement timeout for the MQTT requests in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout, "things-conn-sub-timeout", cfg.ThingsConfig.ThingsConnectionConfig.SubscribeTimeout, "Specify the subscribe timeout for the MQTT requests in milliseconds")
	flagSet.Int64Var(&cfg.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout, "things-conn-unsub-timeout", cfg.ThingsConfig.ThingsConnectionConfig.UnsubscribeTimeout, "Specify the unsubscribe timeout for the MQTT requests in milliseconds")

	// init k8s config
	flagSet.StringVar(&cfg.Orchestration.K8s.Kubeconfig, "k8s-kubeconfig", cfg.Orchestration.K8s.Kubeconfig, "Specify the absolute path to the k8s condiguration")

	// init self update config
	flagSet.BoolVar(&cfg.Orchestration.SelfUpdate.EnableReboot, "self-update-enable-reboot", cfg.Orchestration.SelfUpdate.EnableReboot, "Specify the enable reboot flag to the self update condiguration")
	flagSet.StringVar(&cfg.Orchestration.SelfUpdate.Timeout, "self-update-timeout", cfg.Orchestration.SelfUpdate.Timeout, "Specify the timeout in cron format to wait for completing a self update operation")
	flagSet.StringVar(&cfg.Orchestration.SelfUpdate.RebootTimeout, "self-update-reboot-timeout", cfg.Orchestration.SelfUpdate.RebootTimeout, "Specify the timeout in cron format to wait before a reboot process is initiated after a self update operation")
}

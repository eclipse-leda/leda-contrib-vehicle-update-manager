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
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/spf13/cobra"
)

var (
	cfg = getDefaultInstance()
)

var rootCmd = &cobra.Command{
	Use:               "updatemanagerd",
	Short:             "The Vehicle Update Manager engine",
	Args:              cobra.NoArgs,
	SilenceUsage:      true,
	DisableAutoGenTag: true, // disable displaying auto generation tag in cli docs
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDaemon(cmd)

	},
}

func main() {
	//profile.Profile()
	if reexec.Init() {
		return
	}

	cfgFilePath := parseConfigFilePath()
	if err := loadLocalConfig(cfgFilePath, cfg); err != nil {
		log.ErrorErr(err, "failed to load local configuration provided - will exit")
		os.Exit(1)
	}

	setupCommandFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.ErrorErr(err, "failed to execute root command - will exit")
		os.Exit(1)
	}
}

func runDaemon(cmd *cobra.Command) error {
	initLogger(cfg)
	dumpConfiguration(cfg)

	gwDaemon, err := newDaemon(cfg)
	if err != nil {
		log.ErrorErr(err, "failed to create Update Manager daemon instance")
		return err
	}

	gwDaemon.init()
	if err := util.MkDir("/var/run/updatemanagerd"); err != nil {
		log.WarnErr(err, "error creating root eexec dir")
	}
	runLockFile := path.Join("/var/run/updatemanagerd", string(os.PathSeparator), lockFileName)
	l, lockErr := newRunLock(runLockFile)
	if lockErr == nil {
		err = l.TryLock()
		if err == nil {
			defer l.Unlock()

			err := gwDaemon.start()
			if err != nil {
				log.ErrorErr(err, "failed to start Update Manager daemon instance")
				return err
			}
			log.Debug("successfully started Update Manager daemon instance")

			var signalChan = make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
			select {
			case sig := <-signalChan:
				log.Debug("Received OS SIGNAL >> %d ! Will exit!", sig)
				gwDaemon.stop()
			}
		} else {
			log.ErrorErr(err, "another instance of updatemanagerd is already running")
			return err
		}
	} else {
		log.ErrorErr(lockErr, "unable to create lock file at %s", runLockFile)
		return lockErr
	}
	return nil
}

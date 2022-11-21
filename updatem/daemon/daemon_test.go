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
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/things"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/pkg/testutil"
	"github.com/spf13/cobra"
)

func TestLoadLocalConfig(t *testing.T) {
	cfg := getDefaultInstance()
	t.Run("test_not_existing", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/testdata/config/not-existing.json", cfg)
		if err != nil {
			t.Errorf("null error returned expected for non existing file")
		}
	})
	t.Run("test_is_dir", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/testdata/config/", cfg)
		testutil.AssertError(t, log.NewErrorf("provided configuration path %s is a directory", "../pkg/testutil/testdata/config/"), err)
	})
	t.Run("test_file_empty", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/testdata/config/empty.json", cfg)
		if err != nil {
			t.Errorf("no error expected, only warning")
		}
	})
	t.Run("test_json_invalid", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/testdata/config/invalid.json", cfg)
		testutil.AssertError(t, log.NewError("unexpected end of JSON input"), err)
	})
}

func TestParseConfigFilePath(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	testPath := "/some/path/file"

	t.Run("test_cfg_file_overridden", func(t *testing.T) {
		actualPath := parseConfigFilePath()
		if actualPath != daemonConfigFileDefault {
			t.Error("config file not set to default")
		}
	})
	t.Run("test_cfg_file_default", func(t *testing.T) {
		os.Args = []string{oldArgs[0], fmt.Sprintf("--%s=%s", daemonConfigFileFlagID, testPath)}
		actualPath := parseConfigFilePath()
		if actualPath != testPath {
			t.Error("config file not overridden by environment variable ")
		}
	})
}

// The following test is intended to serve as a check if any of the default configs has changed within a change.
// The test configuration json must be edited in this case - if the change really must be made.
func TestDefaultConfig(t *testing.T) {
	newDefaultConfig := getDefaultInstance()
	defaultConfig := &config{}
	_ = loadLocalConfig("../pkg/testutil/testdata/config/daemon-config.json", defaultConfig)
	if !reflect.DeepEqual(newDefaultConfig, defaultConfig) {
		t.Errorf("default configuration changed: %+v\ngot:%+v", defaultConfig, newDefaultConfig)
	}
}

func TestThingsServiceFeaturesConfig(t *testing.T) {
	local := &config{}
	_ = loadLocalConfig("../pkg/testutil/testdata/config/daemon-things-features-config.json", local)
	testutil.AssertEqual(t, local.ThingsConfig.Features, []string{things.UpdateOrchestratorFeatureID, things.SoftwareUpdatableManifestsFeatureID})
}

func TestExtractOpts(t *testing.T) {
	t.Run("test_extract_things_opts", func(t *testing.T) {
		opts := extractThingsOptions(cfg)
		if opts == nil || len(opts) == 0 {
			t.Error("no things opts after extraction")
		}
	})
	t.Run("test_extract_update_manager_opts", func(t *testing.T) {
		opts := extractUpdateManagerOptions(cfg)
		if opts == nil {
			t.Error("no orchestration opts after extraction")
		}
	})
}

func TestDumpsNoErrors(t *testing.T) {
	cfg := getDefaultInstance()
	dumpConfiguration(cfg)

	t.Run("test_dump_config_nil", func(t *testing.T) {
		dumpConfiguration(nil)
	})
	t.Run("test_dump_config_log_null", func(t *testing.T) {
		logCfg := cfg.Log
		cfg.Log = nil
		dumpConfiguration(cfg)
		cfg.Log = logCfg
	})
	t.Run("test_dump_config_things_null", func(t *testing.T) {
		thingsCfg := cfg.ThingsConfig
		cfg.ThingsConfig = nil
		dumpConfiguration(cfg)
		cfg.ThingsConfig = thingsCfg
	})
	t.Run("test_dump_config_orchestration_null", func(t *testing.T) {
		orchCfg := cfg.Orchestration
		cfg.Orchestration = nil
		dumpConfiguration(cfg)
		cfg.Orchestration = orchCfg
	})
}

func TestSetCommandFlags(t *testing.T) {
	var cmd = &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon(cmd)

		},
	}
	setupCommandFlags(cmd)

	tests := map[string]struct {
		flag         string
		expectedType string
	}{
		"test_flags_log_file": {
			flag:         "log-file",
			expectedType: reflect.String.String(),
		},
		"test_flags_log_level": {
			flag:         "log-level",
			expectedType: reflect.String.String(),
		},
		"test_flags_log_file_size": {
			flag:         "log-file-size",
			expectedType: reflect.Int.String(),
		},
		"test_flags_log_file_count": {
			flag:         "log-file-count",
			expectedType: reflect.Int.String(),
		},
		"test_flags_log_file_max_age": {
			flag:         "log-file-max-age",
			expectedType: reflect.Int.String(),
		},
		"test_flags_log_syslog": {
			flag:         "log-syslog",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_things-home-dir": {
			flag:         "things-home-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-features": {
			flag:         "things-features",
			expectedType: "stringSlice",
		},
		"test_flags_things-conn-broker": {
			flag:         "things-conn-broker",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-keep-alive": {
			flag:         "things-conn-keep-alive",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-disconnect-timeout": {
			flag:         "things-conn-disconnect-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-client-username": {
			flag:         "things-conn-client-username",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-client-password": {
			flag:         "things-conn-client-password",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-connect-timeout": {
			flag:         "things-conn-connect-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-ack-timeout": {
			flag:         "things-conn-ack-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-sub-timeout": {
			flag:         "things-conn-sub-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-unsub-timeout": {
			flag:         "things-conn-unsub-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_orchestration-k8s-kubeconfig": {
			flag:         "k8s-kubeconfig",
			expectedType: reflect.String.String(),
		},
		"test_flags_self-update-enable-reboot": {
			flag:         "self-update-enable-reboot",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_self-update-reboot-timeout": {
			flag:         "self-update-reboot-timeout",
			expectedType: reflect.String.String(),
		},
		"test_flags_self-update-timeout": {
			flag:         "self-update-timeout",
			expectedType: reflect.String.String(),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			flag := cmd.Flag(testCase.flag)
			if flag.Value.Type() != testCase.expectedType {
				t.Errorf("incorrect type: %s for flag %s, expecting: %s", flag.Value.Type(), flag.Name, testCase.expectedType)
			}
			if flag == nil {
				t.Errorf("flag %s, not found", testCase.flag)
			}
		})
	}
}

func TestRunLock(t *testing.T) {
	t.Run("test_lock_create", func(t *testing.T) {
		lock, lockErr := newRunLock(lockFileName)
		if lockErr != nil {
			t.Error("error while creating run lock")
		}
		if lock == nil {
			t.Error("couldn't create lock")
		}
	})

	t.Run("test_lock_try_lock_new_goroutine", func(t *testing.T) {
		lock, _ := newRunLock(lockFileName)
		lockErr := lock.TryLock()
		if lockErr == nil {
			defer func() {
				lock.Unlock()
				_ = os.Remove(lockFileName)
			}()
			go func() {
				secondLock, err := newRunLock(lockFileName)
				if err != nil {
					t.Error("could create second lock")
				}
				lockErrorAlreadyLocked := secondLock.TryLock()
				if lockErrorAlreadyLocked == nil {
					t.Error("run lock locked twice")
				}
			}()
			time.Sleep(1 * time.Second)
		} else {
			t.Error("couldn't create lock")
		}
	})
}

// TODO test the behavior of the daemon towards its services (start, stop), with mocked instanced of GRPC service etc.

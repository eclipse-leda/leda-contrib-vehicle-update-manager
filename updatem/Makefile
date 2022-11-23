# Copyright (c) 2022 Contributors to the Eclipse Foundation
# 
# See the NOTICE file(s) distributed with this work for additional
# information regarding copyright ownership.
#
# This program and the accompanying materials are made available under the
# terms of the Apache License 2.0 which is available at
# https://www.apache.org/licenses/LICENSE-2.0
# 
# SPDX-License-Identifier: Apache-2.0

# DAEMON_BINARY_NAME is the name of binary of daemon.
DAEMON_BINARY_NAME=updatemanagerd

# RESOURCES are the resources to include in each binary
RESOURCES=install.sh uninstall.sh setup_env updatemanagerd.service

# SERVICES defines the project linux services
SERVICES=updatemanagerd.service

# LICENSE_FILE defines the path to the LICENSE file
LICENSE_FILE=../LICENSE

# NOTICE_FILE defines the path to the NOTICE file
NOTICE_FILE=../NOTICE.md

# TESTS_COVERAGE_RES_DIR is the path where the coverage results will be stored
TESTS_COVERAGE_RES_DIR=../

# TESTS_COVERAGE_RES_OUTPUT_FILENAME is the name of the file containing the test coverage results
TESTS_COVERAGE_RES_OUTPUT_FILENAME=updatemanagerd_coverage.out

# RESOURCES_DIR is the directory containing all resources files (e.g. install scripts, service files, etc.)
RESOURCES_DIR=resources

# BINARIES_DEST_DIR is base path used to install edgectr & updatemanagerd
BINARIES_DEST_DIR=/usr/local

# SERVICES_DEST_DIR is base path used to install updatemanagerd service
SERVICES_DEST_DIR=/etc/systemd/system

# GOARCH is the target platform architecture to build for
GOARCH ?= $(shell go env GOARCH)

# GOOS is the target platform OS to build for
GOOS ?= $(shell go env GOOS)

# GOFLAGS defines a readonly mode when building
GOFLAGS=-v -trimpath

# TARGET_PLATFORMS defines the platforms for which the binaries will be built for
TARGET_PLATFORMS=linux/amd64/linux-x86_64 linux/arm/linux-arm linux/arm64/linux-arm64

# INSTALL_ROOT defines the path to the binaries generated per platform
INSTALL_ROOT=bin/install

# ARCHIVES_ROOT defines the path to the archives generated per platform
ARCHIVES_ROOT=bin/archives

# DEST_DIR is base path used to install edgectr & updatemanagerd
DEST_DIR=/usr/local

# API_VERSION is used for daemon API Version in go build.
API_VERSION="0.1"

# VERSION is used for daemon Release Version in go build.
VERSION ?= "v0.1"

# GIT_COMMIT is used for daemon GitCommit in go build.
GIT_COMMIT=$(shell git describe --dirty --always --tags 2> /dev/null || true)

# BUILD_TIME is used for daemon BuildTime in go build.
BUILD_TIME=$(shell date --rfc-3339 s 2> /dev/null | sed -e 's/ /T/')

VERSION_PKG=github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem

DEFAULT_LDFLAGS="-s -w -X ${VERSION_PKG}/version.GitCommit=${GIT_COMMIT} \
		  -X ${VERSION_PKG}/version.ProjectVersion=${VERSION} \
		  -X ${VERSION_PKG}/version.APIVersion=${API_VERSION} \
		  -X ${VERSION_PKG}/version.BuildTime=${BUILD_TIME}"

build-targets: #build updatemanagerd daemon binary for all the target platforms
	@set -e
	@mkdir -p ${INSTALL_ROOT}
	@for platform in ${TARGET_PLATFORMS}; do \
		TARGET_OS=`echo $$platform | cut -d/ -f1`; \
		TARGET_ARCH=`echo $$platform | cut -d/ -f2`; \
		TARGET_INSTALL_DIR=${INSTALL_ROOT}/$$TARGET_OS'_'$$TARGET_ARCH; \
		echo "$@: $$TARGET_INSTALL_DIR"; \
		mkdir -p $$TARGET_INSTALL_DIR; \
		env GOOS=$$TARGET_OS GOARCH=$$TARGET_ARCH go build ${GOFLAGS} -ldflags ${DEFAULT_LDFLAGS} -o $$TARGET_INSTALL_DIR/${DAEMON_BINARY_NAME} github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/daemon || exit 1; \
		install $(addprefix ${RESOURCES_DIR}/,$(RESOURCES)) $$TARGET_INSTALL_DIR ; \
		cp ${LICENSE_FILE} $$TARGET_INSTALL_DIR ; \
		cp ${NOTICE_FILE} $$TARGET_INSTALL_DIR ; \
	done

build-resources-archive:
	@mkdir -p ${ARCHIVES_ROOT}
	@echo "$@: archives for ${RESOURCES_DIR}"
	@tar -czf ${ARCHIVES_ROOT}/updatem.engine.tar.gz -C ${RESOURCES_DIR} .

build-targets-archive: build-targets build-resources-archive ## generates all archives per platforms along with the full archive of all artifacts
	@mkdir -p ${ARCHIVES_ROOT}
	@for platform in ${TARGET_PLATFORMS}; do \
		TARGET_OS=`echo $$platform | cut -d/ -f1`; \
		TARGET_ARCH=`echo $$platform | cut -d/ -f2`; \
		TARGET_PLATFORM_INTERN_CONVEN=`echo $$platform | cut -d/ -f3`; \
		TARGET_DIR=$$TARGET_OS'_'$$TARGET_ARCH; \
		echo "$@: archives for $$TARGET_DIR"; \
		chmod 0755 ${INSTALL_ROOT}/$$TARGET_DIR/${DAEMON_BINARY_NAME} ; \
		chmod 0644 ${INSTALL_ROOT}/$$TARGET_DIR/${SERVICES} ; \
		tar -czf ${ARCHIVES_ROOT}/updatem.engine-$$TARGET_PLATFORM_INTERN_CONVEN.tar.gz -C ${INSTALL_ROOT}/$$TARGET_DIR . ;\
	done
	@echo "$@: compressing all archives ..."
	@cd ${ARCHIVES_ROOT} && tar -czf updatem-package.tar.gz updatem.engine-*.tar.gz

build-local: ## compiles the updatemanagerd daemon binary using the host's OS and ARCH
	@mkdir -p ${INSTALL_ROOT}
	@TARGET_INSTALL_DIR=${INSTALL_ROOT}/${GOOS}_${GOARCH}
	@echo "$@: local target install dir is ${INSTALL_ROOT}/${GOOS}_${GOARCH}"
	@env GOOS=${GOOS} GOARCH=${GOARCH} go build ${GOFLAGS} -ldflags ${DEFAULT_LDFLAGS} -o ${INSTALL_ROOT}/${GOOS}_${GOARCH}/${DAEMON_BINARY_NAME} github.com/eclipse-leda/leda-contrib-vehicle-update-manager/updatem/daemon || exit 1

install: ## install updatemanagerd daemon binary
	@echo $@
	@mkdir -p $(BINARIES_DEST_DIR)/bin
	install $(addprefix ${INSTALL_ROOT}/${GOOS}'_'${GOARCH}/,$(notdir $(DAEMON_BINARY_NAME))) $(BINARIES_DEST_DIR)/bin
	install $(addprefix resources/,$(notdir $(SERVICES))) $(SERVICES_DEST_DIR)
	chmod 0644 $(addprefix $(SERVICES_DEST_DIR)/,$(notdir $(SERVICES)))

test-all: ## executes all unit tests
	@echo $@
	@go test ./... -cover -coverprofile=${TESTS_COVERAGE_RES_DIR}${TESTS_COVERAGE_RES_OUTPUT_FILENAME} || exit 1

uninstall-local: ## uninstall updatemanagerd daemon binary and services
	@echo $@
	@rm -f $(addprefix $(BINARIES_DEST_DIR)/bin/,$(notdir $(DAEMON_BINARY_NAME)))
	@rm -f $(addprefix $(SERVICES_DEST_DIR),$(notdir $(SERVICES)))

clean: ## clean up binaries
	@echo "$@: removed all generated artifacts successfully"
	@echo "$@: cleaning bin/"
	@rm -rf bin/

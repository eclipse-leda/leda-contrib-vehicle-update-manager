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

#!/bin/sh

. ./setup_env

rm -f $BINARIES_DEST_DIR/$DAEMON_BINARY_NAME
echo "- $DAEMON_BINARY_NAME"

rm -f $SERVICES_DEST_DIR/$DAEMON_SERVICE_NAME
echo "- $DAEMON_SERVICE_NAME"

systemctl daemon-reload

FROM --platform=${BUILDPLATFORM} golang:alpine AS build

# Install Alpine Dependencies
RUN apk add --update make git bash

WORKDIR /src
COPY . .
ARG TARGETOS
ARG TARGETARCH

# Set GO env
ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# Build
WORKDIR /src/updatem
RUN make build-local

# Final container
FROM busybox AS bin
ARG TARGETOS
ARG TARGETARCH
COPY --from=build /src/updatem/bin/install/${TARGETOS}_${TARGETARCH}/updatemanagerd /app/

WORKDIR /app
RUN ls -l

# Set config values
ENV CFG_FILE=""

# Set default log vaiues
ENV LOG_FILE=log/update-manager.log
ENV LOG_LEVEL=INFO
ENV LOG_FILE_SIZE=2
ENV LOG_FILE_COUNT=5
ENV LOG_FILE_MAX_AGE=28
ENV LOG_SYSLOG=false

# Set default things values
ENV THINGS_HOME_DIR=/var/lib/updatemanagerd
ENV THINGS_FEATURES=UpdateOrchestrator
ENV THINGS_CONN_BROKER=tcp://127.0.0.1:1883
ENV THINGS_CONN_KEEP_ALIVE=20000
ENV THINGS_CONN_CONNECT_TIMEOUT=30000
ENV THINGS_CONN_DISCONNECT_TIMEOUT=250
ENV THINGS_CONN_ACK_TIMEOUT=15000
ENV THINGS_CONN_SUB_TIMEOUT=15000
ENV THINGS_CONN_UNSUB_TIMEOUT=3000
ENV THINGS_CONN_CLIENT_USERNAME=""
ENV THINGS_CONN_CLIENT_PASSWORD=""

# Set default k8s values
ENV K8S_KUBECONFIG=""

# Set self update values
ENV SELF_UPDATE_ENABLE_REBOOT=false
ENV SELF_UPDATE_TIMEOUT=10m
ENV SELF_UPDATE_REBOOT_TIMEOUT=30s

# Use shell version to pass ENV variables
CMD ./updatemanagerd --things-home-dir=${THINGS_HOME_DIR} --things-features=${THINGS_FEATURES} --things-conn-broker=${THINGS_CONN_BROKER} \
--things-conn-keep-alive=${THINGS_CONN_KEEP_ALIVE} --things-conn-connect-timeout=${THINGS_CONN_CONNECT_TIMEOUT} --things-conn-disconnect-timeout=${THINGS_CONN_DISCONNECT_TIMEOUT} \
--things-conn-ack-timeout=${THINGS_CONN_ACK_TIMEOUT} --things-conn-sub-timeout=${THINGS_CONN_SUB_TIMEOUT} --things-conn-unsub-timeout=${THINGS_CONN_UNSUB_TIMEOUT} \
--things-conn-client-username=${THINGS_CONN_CLIENT_USERNAME} --things-conn-client-password=${THINGS_CONN_CLIENT_PASSWORD} \
--k8s-kubeconfig=${K8S_KUBECONFIG} \
--self-update-enable-reboot=${SELF_UPDATE_ENABLE_REBOOT} --self-update-timeout=${SELF_UPDATE_TIMEOUT} --self-update-reboot-timeout=${SELF_UPDATE_REBOOT_TIMEOUT} \
--log-file=${LOG_FILE} --log-level=${LOG_LEVEL} --log-file-size=${LOG_FILE_SIZE} --log-file-count=${LOG_FILE_COUNT} --log-file-max-age=${LOG_FILE_MAX_AGE} \
--log-syslog=${LOG_SYSLOG}

#!/usr/bin/env bash

set -e

make install-tools

LOCALSTACK_HEALTH_URL='http://localstack:4566/health'
while ! [[ $( curl -s "${LOCALSTACK_HEALTH_URL}" | fgrep '"sqs": "available"' ) ]]; do echo "Localstack not available at $LOCALSTACK_HEALTH_URL, waiting..."; sleep 1; done

apt-get update
apt-get install -y netcat

# fake endpoint so that the 'starter' service can wait on something
nc -lk 0.0.0.0 8080

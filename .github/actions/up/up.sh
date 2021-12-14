#!/usr/bin/env bash

set -xEeuo pipefail # https://vaneyckt.io/posts/safer_bash_scripts_with_set_euxo_pipefail/

make up-daemon

while [[ `docker-compose ps -q starter | xargs docker inspect -f '{{ .State.Status }}'` != 'exited' ]]; do
  echo "Waiting for the Docker environment to be ready..."
  sleep 1
done

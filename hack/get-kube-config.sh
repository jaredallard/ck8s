#!/usr/bin/env bash
# Fetchs a kube config from a locally running k3s cluster

set -e

CONTAINER_NAME=$(docker ps -a | grep _server_1 | awk '{ print $1 }')

if [[ -z "$CONTAINER_NAME" ]]; then
  echo "failed to get container name" 1>&2
  exit 1
fi

docker exec "$CONTAINER_NAME" cat -- /output/kubeconfig.yaml
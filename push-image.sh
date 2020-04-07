#!/bin/sh

if [[ -z "${DOCKER_USERNAME}" ]] || [[ -z "${DOCKER_PASSWORD}" ]]; then
  echo "Please set DOCKER_USERNAME and DOCKER_PASSWORD variables"
  exit 1
fi

REGISTRY_URL="registry.lewiky.com"

docker login --username "${DOCKER_USERNAME}" --password "${DOCKER_PASSWORD}" "${REGISTRY_URL}"

for service in "user" "auth" "match"; do
  docker build -t "$REGISTRY_URL/temple-$service-service" $service
  docker push "$REGISTRY_URL/temple-$service-service"
done

#!/bin/sh

if [[ -z "${DOCKER_USERNAME}" ]] || [[ -z "${DOCKER_PASSWORD}" ]]; then
  echo "Please set DOCKER_USERNAME and DOCKER_PASSWORD variables"
  exit 1
fi


for service in "user" "auth" "match"; do
  docker build -t "registry.lewiky.com/temple-$service-service" $service
  docker push "registry.lewiky.com/temple-$service-service"
done
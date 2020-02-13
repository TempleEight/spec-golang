#!/bin/bash

docker-compose build
docker-compose up
sh kong/configure-kong.sh

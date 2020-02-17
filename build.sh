#!/bin/bash

docker ps -a -q | xargs docker rm -f && docker volume prune -f && docker-compose up --build

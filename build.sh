#!/bin/bash

docker build -t $1 -f $1/Dockerfile $1
docker run -p 80:80 $1

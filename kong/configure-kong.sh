#!/bin/sh

# Add the user service
curl -i -X POST \
  --url http://localhost:8001/services/ \
  --data 'name=user-service' \
  --data 'url=http://user:80/user'

# Add the match service
curl -i -X POST \
  --url http://localhost:8001/services/ \
  --data 'name=match-service' \
  --data 'url=http://match:81/match'

# Add a route for user
curl -i -X POST \
  --url http://localhost:8001/services/user-service/routes \
  --data 'hosts[]=localhost:8000' \
  --data 'paths[]=/api/user'

# Add a route for match
curl -i -X POST \
  --url http://localhost:8001/services/match-service/routes \
  --data 'hosts[]=localhost:8000' \
  --data 'paths[]=/api/match'

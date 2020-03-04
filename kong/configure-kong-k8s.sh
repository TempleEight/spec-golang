#!/bin/sh

# Add the user service
curl -i -X POST \
  --url $KONG_ADMIN/services/ \
  --data 'name=user-service' \
  --data 'url=http://user:80/user'

# Add the match service
curl -i -X POST \
  --url $KONG_ADMIN/services/ \
  --data 'name=match-service' \
  --data 'url=http://match:81/match'

# Add the auth service
curl -i -X POST \
  --url $KONG_ADMIN/services/ \
  --data 'name=auth-service' \
  --data 'url=http://auth:82/auth'

# Add a route for user
curl -i -X POST \
  --url $KONG_ADMIN/services/user-service/routes \
  --data "hosts[]=$KONG_ENTRY" \
  --data 'paths[]=/api/user'

# Add a route for match
curl -i -X POST \
  --url $KONG_ADMIN/services/match-service/routes \
  --data "hosts[]=$KONG_ENTRY" \
  --data 'paths[]=/api/match'

# Add a route for auth
curl -i -X POST \
  --url $KONG_ADMIN/services/auth-service/routes \
  --data "hosts[]=$KONG_ENTRY" \
  --data 'paths[]=/api/auth'

# Require a JWT for the user service
curl -X POST $KONG_ADMIN/services/user-service/plugins \
    --data "name=jwt"\
    --data "config.claims_to_verify=exp"

# Require a JWT for the match service
curl -X POST $KONG_ADMIN/services/match-service/plugins \
    --data "name=jwt"\
    --data "config.claims_to_verify=exp"

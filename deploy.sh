#! /bin/bash
# Deployment script for kube - generates configmaps from SQL init files and provisions everything

GREEN="\033[1;32m"
BLUE="\033[1;34m"
NOCOLOR="\033[0m"

minikube start --vm-driver=virtualbox

echo ${GREEN}

kubectl create secret docker-registry regcred --docker-server=$REG_URL --docker-username=$REG_USERNAME --docker-password=$REG_PASSWORD --docker-email=$REG_EMAIL
kubectl create configmap match-db-config --from-file match-db/init.sql -o=yaml
kubectl create configmap user-db-config --from-file user-db/init.sql -o=yaml
kubectl create configmap auth-db-config --from-file auth-db/init.sql -o=yaml
kubectl create -f kube/kong
kubectl create -f kube/user
kubectl create -f kube/match
kubectl create -f kube/auth

echo ${NOCOLOR}

# Kube takes a few seconds to create the objects
sleep 5

echo Sleeping until pods have started...
# Until all of the pods have status of either "Running" or "Completed"
until ! kubectl get pod | awk '{ if (NR != 1) printf $3 "\n" }' | grep -s -q -E "Running|Completed" --invert
do
  sleep 1
done

# Get the URLs Minikube has assigned to the endpoints
urls=$(minikube service kong --url | head -n 2)

export KONG_ENTRY=$(echo $urls | head -n 1 | cut -d '/' -f 3-)
export KONG_ADMIN=$(echo $urls | tail -n 1)

echo ${BLUE}

echo KONG_ENTRY: $KONG_ENTRY
echo KONG_ADMIN: $KONG_ADMIN

echo ${NOCOLOR}

echo Configuring Kong...

sh kong/configure-kong-k8s.sh

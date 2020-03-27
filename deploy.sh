#! /bin/sh
# Deployment script for kube - generates configmaps from SQL init files and provisions everything

BASEDIR=$(dirname "$0")

GREEN="\033[1;32m"
BLUE="\033[1;34m"
YELLOW="\033[1;33m"
PURPLE="\033[1;34m"
NOCOLOR="\033[0m"

minikube start --vm-driver=virtualbox

echo $GREEN

kubectl create secret docker-registry regcred --docker-server=$REG_URL --docker-username=$REG_USERNAME --docker-password=$REG_PASSWORD --docker-email=$REG_EMAIL
kubectl create configmap match-db-config --from-file "$BASEDIR/match-db/init.sql" -o=yaml
kubectl create configmap user-db-config --from-file "$BASEDIR/user-db/init.sql" -o=yaml
kubectl create configmap auth-db-config --from-file "$BASEDIR/auth-db/init.sql" -o=yaml

for dir in "$BASEDIR/kube/"*
do
  kubectl create -f $dir
done

echo $NOCOLOR

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

echo $BLUE

echo KONG_ENTRY: $KONG_ENTRY
echo KONG_ADMIN: $KONG_ADMIN

echo $YELLOW

echo Configuring Kong...

sleep 1

sh "$BASEDIR/kong/configure-kong-k8s.sh"

echo $PURPLE

echo
echo Done!

echo $NOCOLOR

#! /bin/bash
# Deployment script for kube - generates configmaps from SQL init files and provisions everything
minikube start --vm-driver=virtualbox
kubectl create secret docker-registry regcred --docker-server=$REG_URL --docker-username=$REG_USERNAME --docker-password=$REG_PASSWORD --docker-email=$EMAIL
kubectl create configmap match-db-config --from-file ../match-db/init.sql -o=yaml
kubectl create configmap user-db-config --from-file ../user-db/init.sql -o=yaml
kubectl create -f kong
kubectl create -f user
kubectl create -f match
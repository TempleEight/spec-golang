# Kubernetes Deployment

## _NEW:_ Automated Deployment
_Simply run `source deploy.sh` to do all these steps for you!_

Note: When finished with the cluster, you still need to clean up with

```
$ minikube delete
```

## Manual Deployment

To deploy this project onto a kubernetes cluster follow these steps:

## Cluster setup

First, a Kubernetes cluster is required to deploy to. Instructions aren't provided here for production ready clusters, see the [Kubernetes Documentation](https://kubernetes.io/docs/tasks/) for that. 

For local development, it's recommended to use [Minikube](https://github.com/kubernetes/minikube). Follow the instructions to get it set up.

You can start the minikube deaemon with: `minikube start --vm-driver=virtualbox` (assuming you're using the virtualbox backend).

## Provisioning

The rest of the steps depend on having [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed

### Secrets provision

Some of the containers used in the cluster depend on images in the private Temple docker registry. Kubernetes must be configured with the correct authentication secrets in order to access them.

Run:

```
kubectl create secret docker-registry regcred --docker-server=$REG_URL --docker-username=$REG_USERNAME --docker-password=$REG_PASSWORD --docker-email=$EMAIL
```

Replacing `$REG_URL`, `$REG_USERNAME`, `$REG_PASSWORD`, and `$EMAIL` with the correct values.

This creates the `regcred` secret, used to authenticate k8s with the registry.

### Deploying Infrastructure

All of the Kubernetes configuration YAML files are stored in the `/kube` directory.

Running:

```
kubectl create -f kube/{DIR}
```

Will provision all of the resources described in that directory, so run that command for each subsection of infra you want to provision.

As an example, to deploy all kong infrastructure:

```
kubectl create -f kube/kong
```

### Cleaning up

Once done, clean up the cluster by running

```
minikube delete
```


This will remove all traces of the cluster from your computer.

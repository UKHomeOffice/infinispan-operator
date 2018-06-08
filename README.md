# Infinispan Operator

>**This operator is in WIP state and subject to (breaking) changes.**

This Operator runs an Infinispan cluster based on this blog article: https://blog.infinispan.org/2016/08/running-infinispan-cluster-on-kubernetes.html

The operator itself is built with the: https://github.com/operator-framework/operator-sdk

The main benefit of this Operator is without deeper Kubernetes or Helm and Infinispan operational knowledge you can get an a fully HA Infinispan cluster up and running on your Kubernetes cluster.

Nodes are joining together with JGroups configured to use `KUBE_PING` protocol which finds each Pod running an Infinispan server based on labels and namespace. You can use the default standalone Full HA configuration for Kubernetes.

Infinispan REST and Management endpoints are exposed as `Kubernetes` Services. Infinispan is continuously monitored with by Kubernetes through the built-in Infinispan health checks. 

## Requirements:
 - Install the Operator SDK first: https://github.com/operator-framework/operator-sdk#quick-start

## Usage:

```bash
mkdir -p $GOPATH/src/github.com/banzaicloud
cd $GOPATH/src/github.com/banzaicloud
git clone git@github.com:banzaicloud/infinispan-operator.git
cd infinispan-operator
```

### Get the operator Docker image

#### a. Build the image yourself

```bash
operator-sdk build banzaicloud/infinispan-operator
docker tag banzaicloud/infinispan-operator ${your-operator-image-tag}:latest
docker push ${your-operator-image-tag}:latest
```

#### b. Use the image from Docker Hub

```bash
# No addition steps needed
```

### Install the Kubernetes resources

```bash
kubectl apply -f deploy

kubectl get pods

kubectl get services
```


### The Infinispan Custom Resource

With this YAML template you can install a 3 node Infinispan Cluster easily into your Kubernetes cluster:

```yaml
apiVersion: "infinispan.banzaicloud.com/v1alpha1"
kind: "Infinispan"
metadata:
  name: "infinispan-example"
spec:
  size: 3
```

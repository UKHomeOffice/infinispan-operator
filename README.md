# Infinispan Operator

This Operator runs an Infinispan cluster based on this blog article: https://blog.infinispan.org/2016/08/running-infinispan-cluster-on-kubernetes.html

The operator itself is built with the: https://github.com/operator-framework/operator-sdk

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

#### b. Use the image from Dockerhub

```bash
# No addition steps needed
```

### Install the Kubernetes resources

```bash
kubectl apply -f deploy

kubectl get pods

kubectl get services
```

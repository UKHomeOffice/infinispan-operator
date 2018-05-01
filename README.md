# Infinispan Operator

This Operator runs an Infinispan cluster based on this blog article: https://blog.infinispan.org/2016/08/running-infinispan-cluster-on-kubernetes.html

The operator itself is built with the: https://github.com/operator-framework/operator-sdk

Usage:

```bash
mkdir -p $GOPATH/src/github.com/banzaicloud
cd $GOPATH/src/github.com/banzaicloud
git clone git@github.com:banzaicloud/infinispan-operator.git
cd infinispan-operator

operator-sdk build banzaicloud/infinispan-operator

kubectl apply -f deploy

kubectl get pods

kubectl get services
```

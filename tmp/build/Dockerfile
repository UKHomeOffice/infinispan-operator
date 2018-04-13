FROM alpine:3.6

ADD tmp/_output/bin/infinispan-operator /usr/local/bin/infinispan-operator

RUN adduser -D infinispan-operator
USER infinispan-operator

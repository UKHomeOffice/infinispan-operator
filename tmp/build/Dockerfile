FROM alpine:3.6

ADD tmp/_output/bin/inifinispan-operator /usr/local/bin/inifinispan-operator

RUN adduser -D inifinispan-operator
USER inifinispan-operator

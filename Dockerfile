FROM gsoci.azurecr.io/giantswarm/alpine:3.22.0

RUN apk add --no-cache ca-certificates

ADD ./cluster-apps-operator /cluster-apps-operator

ENTRYPOINT ["/cluster-apps-operator"]

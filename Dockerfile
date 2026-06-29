FROM gsoci.azurecr.io/giantswarm/alpine:3.24.1

RUN apk add --no-cache ca-certificates

ARG TARGETARCH
ADD ./cluster-apps-operator-linux-${TARGETARCH} /cluster-apps-operator

ENTRYPOINT ["/cluster-apps-operator"]

FROM alpine:3.18.5

RUN apk add --no-cache ca-certificates

ADD ./cluster-apps-operator /cluster-apps-operator

ENTRYPOINT ["/cluster-apps-operator"]

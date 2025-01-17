ARG GO_VERSION=1.14
ARG ALPINE_VERSION=3.12

FROM golang:${GO_VERSION} AS builder

RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0

ADD . /go/src/github.com/slipperypenguin/release-notifier
WORKDIR /go/src/github.com/slipperypenguin/release-notifier

#RUN make lint
RUN CGO_ENABLED=0 go build -mod=vendor

FROM alpine:${ALPINE_VERSION}
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/slipperypenguin/release-notifier /bin/
ENTRYPOINT [ "/bin/release-notifier" ]

FROM golang:1.14 as builder

ADD . /go/src/github.com/slipperypenguin/release-tracker
WORKDIR /go/src/github.com/slipperypenguin/release-tracker

RUN make build

FROM alpine:3.12
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/slipperypenguin/release-tracker /bin/
ENTRYPOINT [ "/bin/release-tracker" ]
FROM golang:1.19-alpine AS builder

COPY . /go/src
WORKDIR /go/src
ENV CGO_ENABLED 0
RUN go install -trimpath

FROM alpine:latest

COPY --from=builder /go/bin/markhale.org /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/markhale.org"]

FROM golang:1.17.1 as buildbase
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM buildbase as builder
WORKDIR /app
ADD *.go /app/
RUN export CGO_ENABLED=1 && export GOOS=linux && go build -ldflags "-s -w" -o envoy-eds-server

FROM alpine:3.15
RUN apk --update upgrade && \
    apk add --no-cache libc6-compat && \
    rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/envoy-eds-server .
ENV GIN_MODE=release
ENV HTTP_LISTEN_PORT=8086
ENV GRPC_LISTEN_PORT=8086
ENV NODE_ID=node-0
CMD ["/app/envoy-eds-server"]

FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -installsuffix nocgo -ldflags '-s -w' -v -o fortify

FROM scratch AS minimal
ENV PATH="/usr/local/bin"
WORKDIR /root
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
ENTRYPOINT ["fortify"]

FROM gcr.io/distroless/static-debian12:latest AS distroless
ENV PATH="/usr/local/bin"
WORKDIR /root
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
ENTRYPOINT ["fortify"]

FROM gcr.io/distroless/static-debian12:nonroot AS distroless_nonroot
ENV PATH="/usr/local/bin"
WORKDIR /home/nonroot
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
ENTRYPOINT ["fortify"]

FROM busybox:stable-glibc AS busybox
WORKDIR /root
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
ENTRYPOINT ["fortify"]

FROM alpine:3.21 AS alpine
WORKDIR /root
RUN apk update && apk upgrade && rm -rf /var/cache/apk/*
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
ENTRYPOINT ["fortify"]

FROM debian:stable-slim AS debian
WORKDIR /root
RUN apt-get update && apt-get upgrade -y && apt-get clean && rm -rf /var/lib/apt/lists/*
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
ENTRYPOINT ["fortify"]

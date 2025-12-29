ARG BUILDPLATFORM
ARG SOURCEPATH="."
ARG DEBIAN_BASE="debian:trixie-slim"
ARG ALPINE_BASE="alpine:3.23"
ARG DISTROLESS_BASE="gcr.io/distroless/static-debian12:nonroot"
ARG WOLFI_BASE="cgr.dev/chainguard/wolfi-base:latest"

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG SOURCEPATH
WORKDIR /app
COPY $SOURCEPATH/go.mod $SOURCEPATH/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY $SOURCEPATH/ ./
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
export GOARM=$( [ "$TARGETARCH" = "arm" ] && echo "${TARGETVARIANT}" | tr -d 'v' || echo "" ) && \
GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -v -o fortify

FROM $DEBIAN_BASE AS debian
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/* && \
groupadd -g 65532 nonroot && useradd -u 65532 -g nonroot -s /sbin/nologin -M nonroot
WORKDIR /home/nonroot
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/fortify"]

FROM $ALPINE_BASE AS alpine
RUN apk update && apk upgrade && apk add --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/* && \
addgroup -g 65532 nonroot && adduser -u 65532 -G nonroot -S -D -H nonroot
WORKDIR /home/nonroot
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/fortify"]

FROM $WOLFI_BASE AS wolfi
RUN apk update && apk upgrade
WORKDIR /home/nonroot
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
USER nonroot
ENTRYPOINT ["/usr/local/bin/fortify"]

FROM $DISTROLESS_BASE AS distroless
WORKDIR /home/nonroot
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/fortify"]

FROM scratch AS minimal
COPY --from=builder --chmod=555 /app/fortify /usr/local/bin/
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/fortify"]

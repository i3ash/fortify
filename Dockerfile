FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder
WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -installsuffix nocgo -ldflags '-s -w' -v -o fortify

FROM scratch
WORKDIR /
COPY --from=builder /app/fortify ./
ENTRYPOINT ["/fortify"]

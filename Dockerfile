FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -a -installsuffix cgo -ldflags '-s -w' -o fortify

FROM scratch
WORKDIR /
COPY --from=builder /app/fortify ./
ENTRYPOINT ["/fortify"]

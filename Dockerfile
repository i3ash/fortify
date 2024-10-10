FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -a -installsuffix nocgo -ldflags '-s -w' -v -o fortify

FROM scratch
WORKDIR /
COPY --from=builder /app/fortify ./
ENTRYPOINT ["/fortify"]

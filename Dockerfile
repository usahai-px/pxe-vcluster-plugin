# Build the manager binary
FROM golang:1.20 as builder

# Make sure we use go modules
WORKDIR vcluster

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Install dependencies
RUN go mod download

COPY apis apis
COPY cmd cmd
COPY internal internal

# Build cmd
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o /plugin ./cmd/main.go

# we use alpine for easier debugging
FROM alpine

# Set root path as working directory
WORKDIR /

COPY --from=builder /plugin .

ENTRYPOINT ["/plugin"]
# build image #
FROM golang:1.16.3 AS build

WORKDIR /src

# Don't use the proxy
ENV GOPRIVATE=*

# Copy over the module stuff first
COPY go.mod go.sum ./

# Download the modules to cache them.
RUN go mod download

# Copy over the rest of the files.
COPY . .

# Run the unit tests.
RUN go test -race ./...

# Run go vet
RUN go vet ./...

# Run static check
RUN go run honnef.co/go/tools/cmd/staticcheck ./...

# Run err check
RUN go run github.com/kisielk/errcheck ./...

# Make sure go fmt was run.
# If this fails, you should configure your editor to run `gofmt` on save.
RUN test -z $(go fmt ./...)

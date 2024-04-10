# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.18 as builder

# Create and change to the app directory.
WORKDIR /app

# Copy local code to the container image.
COPY . .

# Resolve application dependencies.
# Using go mod with Go 1.11 modules or later.
RUN go mod download
RUN go mod verify

# Build the binary.
# -o myapp specifies the output binary name "myapp".
RUN CGO_ENABLED=0 GOOS=linux go build -v -o myapp

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/
# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
FROM debian:buster-slim

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/myapp /myapp

# Set the binary as the entrypoint of the container.
ENTRYPOINT ["/myapp"]

# Service listens on port 8090.
EXPOSE 8090

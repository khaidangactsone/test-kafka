FROM golang:1.20

WORKDIR /app

RUN apt-get update && apt-get install -y librdkafka-dev

COPY . .

# Fetch dependencies.
# Using go get.
RUN go get -d -v ./...
# Using go mod.
RUN go mod tidy

# Build the Go app
RUN go build -o main .

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/myapp /myapp

# Copy the configuration file into the container image
COPY getting-started.properties /getting-started.properties

# Run the web service on container startup.
CMD ["/myapp"]
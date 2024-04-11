# Start from the official Go image to build our application
FROM golang:1.20.4 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

COPY . .

ENV GOPROXY=direct

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

######## Start a new stage from scratch #######
# This is the final stage where the executable is run in a clean image
FROM scratch  

COPY --from=builder /app/myapp .

# Expose port 8090 to the outside world
EXPOSE 8090

CMD ["go", "run", "main.go", "getting-started.properties"]

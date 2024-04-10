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

# Command to run the executable
CMD ["./main", "config.txt"]
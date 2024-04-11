# Use the official Go image as a parent image
FROM golang:1.20.4

# Set the working directory inside the container

# Bypass the proxy for module downloads
ENV GOPROXY=direct

WORKDIR /app

# Copy the local configuration file to the container
COPY . .

# Download all the dependencies
RUN go mod download

# Expose the port the app runs on
EXPOSE 8090

# Run the application with "go run", passing "getting-started.properties" as an argument
CMD ["go", "run", "main.go", "getting-started.properties"]

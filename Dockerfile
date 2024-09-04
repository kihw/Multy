# Use an Ubuntu-based image
FROM ubuntu:20.04

# Set environment variables for Go
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Set noninteractive mode for apt-get
ENV DEBIAN_FRONTEND=noninteractive

# Install necessary packages
RUN apt-get update && apt-get install -y \
    wget \
    libx11-dev \
    libx11-xcb-dev \
    libxkbcommon-dev \
    libxkbcommon-x11-dev \
    libxext-dev \
    x11proto-core-dev \
    x11proto-input-dev \
    x11proto-record-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Download and install Go 1.20 (or update to a newer stable version)
RUN wget https://go.dev/dl/go1.18.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz && \
    rm go1.18.linux-amd64.tar.gz

# Set the working directory in the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies first
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .
RUN go get .
# Build the Go application
RUN go run main.go

# Expose the port the app runs on (replace 8080 with your actual port if necessary)
EXPOSE 8080

# Command to run the application
CMD ["./main"]

# Start from a golang base image
FROM golang:1.19 as builder

# Set the current working directory inside the container
WORKDIR /go/src/app/

# Copy the source code
COPY . .

# Download all dependencies.
RUN go mod download

# Build the Go app with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o server

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./server"]

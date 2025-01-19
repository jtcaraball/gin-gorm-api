# Start from golang base image
FROM golang:1.23.4-alpine as builder
# Install swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest
# Working directory
WORKDIR /src
# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies
RUN go mod download
# Copy everythings
COPY . .
# Update swagger docs
RUN swag init
# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o main .

# Start a new stage from scratch
FROM alpine:latest
# Working directory
WORKDIR /root/
# Copy the pre-built binary file from the previous stage.
COPY --from=builder /src/main .
# Expose port 8080 to the outside world
EXPOSE 8080
#Command to run the executable
CMD ["./main"]

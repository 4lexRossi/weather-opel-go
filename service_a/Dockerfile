# Use GoLang official image
FROM golang:1.22.5

# Set working directory
WORKDIR /app

# Copy files
COPY . .

# Download dependencies
RUN go mod tidy

# Build the application
RUN go build -o service-a main.go

# Expose port
EXPOSE 8081

# Command to run the service
CMD ["./service-a"]

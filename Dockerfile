# Base stage for building the application
FROM golang:1.23-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server \ 
  && CGO_ENABLED=0 GOOS=linux go build -o /client ./cmd/client

# Final stage for server image
FROM alpine:3.21 AS server
WORKDIR /root/
COPY --from=build /server .
EXPOSE 8080 
CMD ["./server"]

# Final stage for client image
FROM alpine:3.21 AS client
WORKDIR /root/
COPY --from=build /client .
CMD ["./client"]

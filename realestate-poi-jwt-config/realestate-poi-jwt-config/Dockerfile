FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download
#RUN go mod tidy

# Copy the rest of the application files
COPY . .

# Build the Go application
RUN go build -o main .


FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
# Set working directory in the new container
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/web ./web
COPY --from=builder /app/pkg/configs ./pkg/configs
# COPY --from=builder /app/pkg/database ./pkg/database


EXPOSE 8000

# Run the application
CMD ["./main"]

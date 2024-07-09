# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/url-shortener .
EXPOSE 8000
CMD ["./url-shortener"]

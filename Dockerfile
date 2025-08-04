# Dockerfile for autobackcom Go app
FROM golang:1.24.5-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o autobackcom ./cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/autobackcom .
COPY .env .
EXPOSE 8080
CMD ["./autobackcom"]

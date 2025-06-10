FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod .
RUN go mod download
COPY . .
RUN go mod tidy
RUN go build -o server

FROM alpine
WORKDIR /app
COPY --from=builder /app/server .
COPY static ./static
EXPOSE 8080
CMD ["./server"]

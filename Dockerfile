FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o server

FROM alpine
WORKDIR /app
COPY --from=builder /app/server .
COPY static ./static
EXPOSE 8080
CMD ["./server"]

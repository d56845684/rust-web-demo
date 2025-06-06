# 建立 Builder
FROM rust:1.87.0-slim-bullseye AS builder

WORKDIR /app
COPY . .

# Enable immediate exit on any error
RUN set -e

# Fetch dependencies and check for errors
RUN cargo fetch
# RUN cargo check --message-format=json 1>&2 || exit 1
RUN cargo build --release

# 建立執行環境（較小）
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/target/release/rust-web-demo .
COPY static ./static
EXPOSE 8080
CMD ["./rust-web-demo"]
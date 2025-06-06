# 建立 Builder
FROM rust:1.87.0-slim-bullseye AS builder

WORKDIR /app
COPY . .
# 預先下載依賴 (提高建置效率)
RUN cargo fetch
RUN cargo check
RUN cargo build --release

# 建立執行環境（較小）
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/target/release/rust-web-demo .
COPY static ./static
EXPOSE 8080
CMD ["./rust-web-demo"]
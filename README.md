# Rust Web Demo

這是一個使用 Rust 和 Actix-web 框架構建的簡單待辦事項（Todo）應用程式。該應用程式使用 PostgreSQL 作為數據庫，並提供了完整的 CRUD（創建、讀取、更新、刪除）操作。

## 功能特點

- RESTful API 端點
- PostgreSQL 數據庫集成
- Docker 容器化支持
- 完整的日誌記錄
- UUID 作為主鍵
- 自動化的構建和運行腳本

## 技術棧

- Rust
- Actix-web
- PostgreSQL
- Docker
- Tokio (異步運行時)
- Deadpool (數據庫連接池)
- Serde (序列化/反序列化)
- UUID

## 系統要求

- Rust (最新穩定版)
- Docker
- Docker Compose
- PostgreSQL (如果不在 Docker 中運行)

## 快速開始

1. 克隆專案：
```bash
git clone <repository-url>
cd rust-web-demo
```

2. 運行應用程式：
```bash
./auto_build.sh
./auto_run.sh
```

這個腳本會：
- 停止並移除現有的容器
- 構建並啟動 PostgreSQL 容器
- 構建並運行 Rust 應用程式

## API 端點

- `GET /todos` - 獲取所有待辦事項
- `POST /todos` - 創建新的待辦事項
- `PUT /todos/{id}` - 更新待辦事項
- `POST /todos/{id}/toggle` - 切換待辦事項的完成狀態
- `DELETE /todos/{id}` - 刪除待辦事項
- `GET /test-db` - 測試數據庫連接
- `POST /api/login` - 使用者登入並取得 JWT

登入頁面可透過 `/login` 進入，所有 Todo API 需在 `Authorization` header 中附帶 `Bearer <token>`。

## 數據庫結構

```sql
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT false
);
```

## 開發

### 構建

```bash
./auto_build.sh
```

### 環境變量

- `RUST_LOG` - 設置日誌級別（默認：info）
- 數據庫配置（在代碼中設置）：
  - 主機：postgres
  - 端口：5432
  - 數據庫名：rust_demo
  - 用戶名：rust_user
  - 密碼：rust_password

## 項目結構

```
rust-web-demo/
├── src/
│   └── main.rs          # 主應用程式代碼
├── static/              # 靜態文件
├── Cargo.toml          # Rust 依賴配置
├── Dockerfile          # Docker 構建配置
├── auto_run.sh         # 自動運行腳本
└── auto_build.sh       # 自動構建腳本
```

## 日誌

應用程式使用 `log` 和 `env_logger` 進行日誌記錄。所有 API 操作都會記錄詳細信息，包括：
- 請求處理
- 數據庫操作
- 錯誤信息

# Go Web Demo

這是一個使用 Go 和 Gin 框架構建的簡單待辦事項（Todo）應用程式，
使用 PostgreSQL 作為資料庫並提供完整的 CRUD 操作。

## 功能特點

- RESTful API 端點
- PostgreSQL 數據庫整合
- Docker 容器化支援
- 簡易日誌記錄
- UUID 作為主鍵
- 自動化構建與啟動腳本

## 技術棧

- Go
- Gin
- PostgreSQL
- Docker
- JWT 驗證
- database/sql 與 pq 驅動

## 系統要求

- Go (最新穩定版)
- Docker
- Docker Compose
- PostgreSQL（若未在 Docker 中執行）

## 快速開始

1. 克隆專案：
```bash
git clone <repository-url>
cd rust-web-demo
```

2. 啟動應用程式：
```bash
./auto_build.sh
./auto_run.sh
```

腳本會：
- 停止並移除現有容器
- 建立並啟動 PostgreSQL 容器
- 建構並執行 Go 應用程式

## API 端點

- `GET /todos` - 取得所有待辦事項
- `POST /todos` - 建立新的待辦事項
- `PUT /todos/{id}` - 更新待辦事項
- `POST /todos/{id}/toggle` - 切換待辦事項完成狀態
- `DELETE /todos/{id}` - 刪除待辦事項
- `GET /test-db` - 測試資料庫連線
- `POST /api/login` - 使用者登入取得 JWT
- `POST /api/register` - 使用者註冊

登入頁面位於 `/login`，所有 Todo API 需在 `Authorization`
header 中附帶 `Bearer <token>`。

## 數據庫結構

```sql
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT false,
    username TEXT NOT NULL REFERENCES users(username)
);
```

## 開發

### 構建

```bash
./auto_build.sh
```

### 環境變量

- 資料庫配置（程式碼中設定）：
  - 主機：postgres
  - 端口：5432
  - 數據庫名：go_demo
  - 用戶名：go_user
  - 密碼：go_password

## 項目結構

```
rust-web-demo/
├── main.go          # 主應用程式程式碼
├── go.mod           # Go 依賴配置
├── static/          # 靜態檔案
├── Dockerfile       # Docker 構建設定
├── auto_run.sh      # 自動執行腳本
└── auto_build.sh    # 自動構建腳本
```

## 日誌

應用程式使用標準庫 `log` 進行簡易日誌記錄，
所有 API 操作會記錄基本資訊與錯誤訊息。

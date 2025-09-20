# Go Web Demo

這是一個使用 Go、Gin 與 React 建構的簡單待辦事項（Todo）應用程式，
使用 PostgreSQL 作為資料庫並提供完整的 CRUD 操作。

## 功能特點

- React SPA 前端
- RESTful API 端點
- PostgreSQL 數據庫整合
- Docker 容器化支援
- 簡易日誌記錄
- UUID 作為主鍵
- 自動化構建與啟動腳本
- JWT 驗證中間件

## 技術棧

- Go
- Gin
- PostgreSQL
- Docker & Docker Compose
- JWT 驗證
- database/sql 與 pq 驅動

## 系統要求

- Go (最新穩定版)
- Docker
- Docker Compose
- Node.js (若要在本地直接執行前端)

## 快速開始

1. 克隆專案：
```bash
git clone <repository-url>
cd rust-web-demo
```

2. 透過 Docker Compose 啟動整個系統：
```bash
docker compose up --build
```

這會啟動三個服務：

- `postgres`：PostgreSQL 15 資料庫
- `backend`：Go API 服務，執行於 `http://localhost:8080`
- `frontend`：React 應用，執行於 `http://localhost:3000`

3. （可選）本地啟動前端 React 開發伺服器：

```bash
cd frontend
npm install
npm run dev -- --host
```

預設會從 `VITE_API_BASE_URL` 環境變數讀取 API 位址，若未設定則使用 `http://localhost:8080`。

## API 端點

- `GET /todos` - 取得所有待辦事項
- `POST /todos` - 建立新的待辦事項
- `PUT /todos/{id}` - 更新待辦事項
- `POST /todos/{id}/toggle` - 切換待辦事項完成狀態
- `DELETE /todos/{id}` - 刪除待辦事項
- `GET /test-db` - 測試資料庫連線
- `POST /api/login` - 使用者登入取得 JWT
- `POST /api/register` - 使用者註冊

所有 Todo API 需在 `Authorization` header 中附帶 `Bearer <token>`。
使用 Docker Compose 啟動後，可於瀏覽器造訪 `http://localhost:3000` 操作完整介面。

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

### 環境變數

| 名稱 | 作用 | 預設值 |
| --- | --- | --- |
| `DATABASE_URL` | PostgreSQL 連線字串 | `postgres://go_user:go_password@postgres:5432/go_demo?sslmode=disable` |
| `CORS_ALLOWED_ORIGINS` | 允許跨來源的前端網址（逗號分隔） | `*` |
| `VITE_API_BASE_URL` | 前端請求 API 的基底位址 | `http://localhost:8080` |

前端在 Docker 中建置時會將 `VITE_API_BASE_URL` 當成建置時常數，若需指向其他位址可以在 `docker compose` 中調整 build args。

## 項目結構

```
rust-web-demo/
├── main.go          # 主應用程式程式碼
├── go.mod           # Go 依賴配置
├── go.sum
├── static/          # 靜態檔案
├── Dockerfile       # 後端 Docker 構建設定
├── docker-compose.yml
├── auto_run.sh      # 自動執行腳本
├── auto_build.sh    # 自動構建腳本
└── frontend/        # React 前端專案
```

## 日誌

應用程式使用標準庫 `log` 進行簡易日誌記錄，
所有 API 操作會記錄基本資訊與錯誤訊息。

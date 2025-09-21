package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

const (
	defaultJWTSecret   = "secret_key"
	defaultDatabaseURL = "postgres://go_user:go_password@postgres:5432/go_demo?sslmode=disable"
	defaultServerPort  = "8080"
	defaultServerHost  = "0.0.0.0"
)

var (
	db            *sql.DB
	jwtSecret     string
	serverAddress string
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLogLevel = LevelInfo
	logLevelNames   = map[string]LogLevel{
		"debug": LevelDebug,
		"info":  LevelInfo,
		"warn":  LevelWarn,
		"error": LevelError,
	}
)

func configureLogging() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	levelName := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if levelName == "" {
		levelName = "info"
	}
	if level, ok := logLevelNames[levelName]; ok {
		currentLogLevel = level
		logDebug("log level configured", "level", levelName)
	} else {
		logWarn("invalid LOG_LEVEL value, defaulting to info", "value", levelName)
		currentLogLevel = LevelInfo
	}
}

func logWithLevel(level LogLevel, label string, message string, keyvals ...interface{}) {
	if currentLogLevel > level {
		return
	}
	builder := strings.Builder{}
	builder.WriteString("[")
	builder.WriteString(label)
	builder.WriteString("] ")
	builder.WriteString(message)
	if len(keyvals) > 0 {
		builder.WriteString(" | ")
		builder.WriteString(formatKeyvals(keyvals...))
	}
	log.Println(builder.String())
}

func formatKeyvals(keyvals ...interface{}) string {
	if len(keyvals) == 0 {
		return ""
	}
	parts := make([]string, 0, len(keyvals)/2)
	for i := 0; i+1 < len(keyvals); i += 2 {
		key := fmt.Sprint(keyvals[i])
		value := fmt.Sprint(keyvals[i+1])
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	if len(keyvals)%2 == 1 {
		parts = append(parts, fmt.Sprint(keyvals[len(keyvals)-1]))
	}
	return strings.Join(parts, " ")
}

func logDebug(message string, keyvals ...interface{}) {
	logWithLevel(LevelDebug, "DEBUG", message, keyvals...)
}

func logInfo(message string, keyvals ...interface{}) {
	logWithLevel(LevelInfo, "INFO", message, keyvals...)
}

func logWarn(message string, keyvals ...interface{}) {
	logWithLevel(LevelWarn, "WARN", message, keyvals...)
}

func logError(message string, keyvals ...interface{}) {
	logWithLevel(LevelError, "ERROR", message, keyvals...)
}

func logFatal(message string, keyvals ...interface{}) {
	logError(message, keyvals...)
	os.Exit(1)
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			logWarn("unable to open .env file", "path", path, "error", err)
		}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			logWarn("invalid line in .env file", "path", path, "line", lineNumber)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		if key == "" {
			logWarn("missing key in .env file", "path", path, "line", lineNumber)
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			logWarn("unable to set environment variable from .env", "key", key, "error", err)
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		logWarn("error reading .env file", "path", path, "error", err)
	} else {
		logDebug("loaded .env file", "path", path)
	}
}

func initConfig() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = defaultJWTSecret
		logWarn("JWT_SECRET not provided, using default value")
	}

	serverAddress = determineServerAddress()
	logInfo("server address configured", "address", serverAddress)
}

func determineServerAddress() string {
	if addr := strings.TrimSpace(os.Getenv("SERVER_ADDRESS")); addr != "" {
		return addr
	}
	port := strings.TrimSpace(os.Getenv("SERVER_PORT"))
	if port == "" {
		port = strings.TrimSpace(os.Getenv("PORT"))
	}
	if port == "" {
		port = defaultServerPort
	}
	if strings.Contains(port, ":") {
		return port
	}
	host := strings.TrimSpace(os.Getenv("SERVER_HOST"))
	if host == "" {
		host = defaultServerHost
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func sanitizeDatabaseURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if parsed.User != nil {
		username := parsed.User.Username()
		if _, hasPassword := parsed.User.Password(); hasPassword {
			parsed.User = url.UserPassword(username, "****")
		} else {
			parsed.User = url.User(username)
		}
	}
	return parsed.String()
}

func initDB() error {
	var err error
	conn := os.Getenv("DATABASE_URL")
	if conn == "" {
		logWarn("DATABASE_URL not provided, using default connection string")
		conn = defaultDatabaseURL
	} else {
		logDebug("DATABASE_URL loaded from environment", "url", sanitizeDatabaseURL(conn))
	}
	logInfo("connecting to database", "url", sanitizeDatabaseURL(conn))
	db, err = sql.Open("postgres", conn)
	if err != nil {
		logError("failed to open database connection", "error", err)
		return err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if err = db.Ping(); err != nil {
		logError("database ping failed", "error", err)
		return err
	}
	logInfo("database connection established")
	// setup tables
	stmts := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
		`CREATE TABLE IF NOT EXISTS users (
            username TEXT PRIMARY KEY,
            password TEXT NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS todos (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            title TEXT NOT NULL,
            done BOOLEAN NOT NULL DEFAULT false,
            username TEXT NOT NULL REFERENCES users(username)
        )`,
		`INSERT INTO users (username, password) VALUES ('admin', 'password') ON CONFLICT (username) DO NOTHING`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			logError("failed to execute setup statement", "error", err)
			return err
		}
	}
	logInfo("database schema ensured")
	return nil
}

type Todo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Done     bool   `json:"done"`
	Username string `json:"username"`
}

type NewTodo struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type UpdateTodo struct {
	Title *string `json:"title"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// authMiddleware checks the Authorization header for a valid JWT
// and stores the username in the request context.
func authMiddleware(c *gin.Context) {
	username, err := authorize(c)
	if err != nil {
		logWarn("authorization failed", "path", c.FullPath(), "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	logDebug("authorization successful", "path", c.FullPath(), "username", username)
	c.Set("username", username)
	c.Next()
}

func authorize(c *gin.Context) (string, error) {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil {
			logDebug("failed to parse JWT", "error", err)
			return "", errors.New("unauthorized")
		}
		if token.Valid {
			logDebug("JWT validated", "username", claims.Subject)
			return claims.Subject, nil
		}
		logDebug("JWT token invalid")
		return "", errors.New("unauthorized")
	}
	if auth != "" {
		logDebug("unsupported authorization header format")
	}
	return "", errors.New("unauthorized")
}

func allowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		logWarn("CORS_ALLOWED_ORIGINS not set, allowing all origins")
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		logWarn("CORS_ALLOWED_ORIGINS produced no valid entries, allowing all origins")
		return []string{"*"}
	}
	logInfo("CORS allowed origins configured", "origins", strings.Join(origins, ","))
	return origins
}

func corsMiddleware() gin.HandlerFunc {
	origins := allowedOrigins()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if len(origins) == 1 && origins[0] == "*" {
			if origin == "" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		} else if isOriginAllowed(origin, origins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}
	for _, item := range allowed {
		if strings.EqualFold(item, origin) {
			return true
		}
	}
	return false
}

func register(c *gin.Context) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		logWarn("register: invalid payload", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}
	logInfo("register: attempting to create user", "username", req.Username)
	res, err := db.Exec(`INSERT INTO users (username, password) VALUES ($1,$2) ON CONFLICT (username) DO NOTHING`, req.Username, req.Password)
	if err != nil {
		logError("register: failed to insert user", "username", req.Username, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		logInfo("register: user already exists", "username", req.Username)
	} else {
		logInfo("register: user created", "username", req.Username)
	}
	c.Status(http.StatusOK)
}

func login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		logWarn("login: invalid payload", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}
	logInfo("login: attempt", "username", req.Username)
	var password string
	err := db.QueryRow(`SELECT password FROM users WHERE username=$1`, req.Username).Scan(&password)
	if err != nil {
		logWarn("login: user lookup failed", "username", req.Username, "error", err)
		c.Status(http.StatusUnauthorized)
		return
	}
	if password != req.Password {
		logWarn("login: invalid credentials", "username", req.Username)
		c.Status(http.StatusUnauthorized)
		return
	}
	claims := jwt.RegisteredClaims{Subject: req.Username, ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour))}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		logError("login: failed to sign JWT", "username", req.Username, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	logInfo("login: success", "username", req.Username)
	c.JSON(http.StatusOK, LoginResponse{Token: tokenStr})
}

func getTodos(c *gin.Context) {
	username := c.GetString("username")
	logDebug("todos: fetching", "username", username)
	rows, err := db.Query(`SELECT id::text, title, done, username FROM todos WHERE username=$1`, username)
	if err != nil {
		logError("todos: failed to query", "username", username, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.Username); err != nil {
			logError("todos: failed to scan row", "username", username, "error", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		logError("todos: iteration error", "username", username, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	logInfo("todos: fetched", "username", username, "count", len(todos))
	c.JSON(http.StatusOK, todos)
}

func addTodo(c *gin.Context) {
	username := c.GetString("username")
	var n NewTodo
	if err := c.BindJSON(&n); err != nil {
		logWarn("todos: add invalid payload", "username", username, "error", err)
		c.Status(http.StatusBadRequest)
		return
	}
	var t Todo
	err := db.QueryRow(`INSERT INTO todos (title, done, username) VALUES ($1,$2,$3) RETURNING id::text, title, done, username`, n.Title, n.Done, username).Scan(&t.ID, &t.Title, &t.Done, &t.Username)
	if err != nil {
		logError("todos: failed to insert", "username", username, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	logInfo("todos: added", "username", username, "todo_id", t.ID)
	c.JSON(http.StatusOK, t)
}

func toggleTodo(c *gin.Context) {
	username := c.GetString("username")
	id := c.Param("id")
	var t Todo
	err := db.QueryRow(`UPDATE todos SET done = NOT done WHERE id::text=$1 AND username=$2 RETURNING id::text, title, done, username`, id, username).Scan(&t.ID, &t.Title, &t.Done, &t.Username)
	if err != nil {
		logWarn("todos: toggle failed", "username", username, "todo_id", id, "error", err)
		c.Status(http.StatusNotFound)
		return
	}
	logInfo("todos: toggled", "username", username, "todo_id", t.ID, "done", t.Done)
	c.JSON(http.StatusOK, t)
}

func deleteTodo(c *gin.Context) {
	username := c.GetString("username")
	id := c.Param("id")
	res, err := db.Exec(`DELETE FROM todos WHERE id::text=$1 AND username=$2`, id, username)
	if err != nil {
		logError("todos: delete failed", "username", username, "todo_id", id, "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		logWarn("todos: delete no rows", "username", username, "todo_id", id)
		c.Status(http.StatusNotFound)
		return
	}
	logInfo("todos: deleted", "username", username, "todo_id", id)
	c.Status(http.StatusNoContent)
}

func updateTodo(c *gin.Context) {
	username := c.GetString("username")
	id := c.Param("id")
	var upd UpdateTodo
	if err := c.BindJSON(&upd); err != nil {
		logWarn("todos: update invalid payload", "username", username, "todo_id", id, "error", err)
		c.Status(http.StatusBadRequest)
		return
	}
	if upd.Title == nil {
		logWarn("todos: update missing title", "username", username, "todo_id", id)
		c.Status(http.StatusBadRequest)
		return
	}
	var t Todo
	err := db.QueryRow(`UPDATE todos SET title=$1 WHERE id::text=$2 AND username=$3 RETURNING id::text, title, done, username`, *upd.Title, id, username).Scan(&t.ID, &t.Title, &t.Done, &t.Username)
	if err != nil {
		logWarn("todos: update failed", "username", username, "todo_id", id, "error", err)
		c.Status(http.StatusNotFound)
		return
	}
	logInfo("todos: updated", "username", username, "todo_id", t.ID)
	c.JSON(http.StatusOK, t)
}

func testDB(c *gin.Context) {
	if err := db.Ping(); err != nil {
		logError("test-db: ping failed", "error", err)
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	logInfo("test-db: ping successful")
	c.JSON(http.StatusOK, "Database connection successful!")
}

func main() {
	loadEnvFile(".env")
	configureLogging()
	initConfig()

	if err := initDB(); err != nil {
		logFatal("failed to initialize database", "error", err)
	}

	if mode := strings.TrimSpace(os.Getenv("GIN_MODE")); mode != "" {
		gin.SetMode(mode)
		logInfo("gin mode configured", "mode", mode)
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) { c.File("./static/index.html") })
	r.GET("/login", func(c *gin.Context) { c.File("./static/login.html") })
	r.GET("/register", func(c *gin.Context) { c.File("./static/register.html") })

	r.POST("/api/register", register)
	r.POST("/api/login", login)

	todos := r.Group("/todos", authMiddleware)
	todos.GET("", getTodos)
	todos.POST("", addTodo)
	todos.PUT("/:id", updateTodo)
	todos.POST("/:id/toggle", toggleTodo)
	todos.DELETE("/:id", deleteTodo)

	r.GET("/test-db", testDB)

	logInfo("starting server", "address", serverAddress)
	if err := r.Run(serverAddress); err != nil {
		logFatal("server stopped", "error", err)
	}
}

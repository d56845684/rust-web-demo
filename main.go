package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

const jwtSecret = "secret_key"

var db *sql.DB

func initDB() error {
	var err error
	conn := "postgres://rust_user:rust_password@postgres:5432/rust_demo?sslmode=disable"
	db, err = sql.Open("postgres", conn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if err = db.Ping(); err != nil {
		return err
	}
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
			return err
		}
	}
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

func authorize(c *gin.Context) (string, error) {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err == nil && token.Valid {
			return claims.Subject, nil
		}
	}
	return "", errors.New("unauthorized")
}

func register(c *gin.Context) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if _, err := db.Exec(`INSERT INTO users (username, password) VALUES ($1,$2) ON CONFLICT (username) DO NOTHING`, req.Username, req.Password); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var password string
	err := db.QueryRow(`SELECT password FROM users WHERE username=$1`, req.Username).Scan(&password)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	if password != req.Password {
		c.Status(http.StatusUnauthorized)
		return
	}
	claims := jwt.RegisteredClaims{Subject: req.Username, ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour))}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, LoginResponse{Token: tokenStr})
}

func getTodos(c *gin.Context) {
	username, err := authorize(c)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	rows, err := db.Query(`SELECT id::text, title, done, username FROM todos WHERE username=$1`, username)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.Username); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		todos = append(todos, t)
	}
	c.JSON(http.StatusOK, todos)
}

func addTodo(c *gin.Context) {
	username, err := authorize(c)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	var n NewTodo
	if err := c.BindJSON(&n); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var t Todo
	err = db.QueryRow(`INSERT INTO todos (title, done, username) VALUES ($1,$2,$3) RETURNING id::text, title, done, username`, n.Title, n.Done, username).Scan(&t.ID, &t.Title, &t.Done, &t.Username)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, t)
}

func toggleTodo(c *gin.Context) {
	username, err := authorize(c)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	id := c.Param("id")
	var t Todo
	err = db.QueryRow(`UPDATE todos SET done = NOT done WHERE id::text=$1 AND username=$2 RETURNING id::text, title, done, username`, id, username).Scan(&t.ID, &t.Title, &t.Done, &t.Username)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, t)
}

func deleteTodo(c *gin.Context) {
	username, err := authorize(c)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	id := c.Param("id")
	res, err := db.Exec(`DELETE FROM todos WHERE id::text=$1 AND username=$2`, id, username)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusNoContent)
}

func updateTodo(c *gin.Context) {
	username, err := authorize(c)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	id := c.Param("id")
	var upd UpdateTodo
	if err := c.BindJSON(&upd); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if upd.Title == nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var t Todo
	err = db.QueryRow(`UPDATE todos SET title=$1 WHERE id::text=$2 AND username=$3 RETURNING id::text, title, done, username`, *upd.Title, id, username).Scan(&t.ID, &t.Title, &t.Done, &t.Username)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, t)
}

func testDB(c *gin.Context) {
	if err := db.Ping(); err != nil {
		c.String(http.StatusInternalServerError, "Database error: %v", err)
		return
	}
	c.JSON(http.StatusOK, "Database connection successful!")
}

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("db init: %v", err)
	}

	r := gin.Default()
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) { c.File("./static/index.html") })
	r.GET("/login", func(c *gin.Context) { c.File("./static/login.html") })
	r.GET("/register", func(c *gin.Context) { c.File("./static/register.html") })

	r.POST("/api/register", register)
	r.POST("/api/login", login)

	r.GET("/todos", getTodos)
	r.POST("/todos", addTodo)
	r.PUT("/todos/:id", updateTodo)
	r.POST("/todos/:id/toggle", toggleTodo)
	r.DELETE("/todos/:id", deleteTodo)

	r.GET("/test-db", testDB)

	log.Println("Starting server on :8080")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal(err)
	}
}

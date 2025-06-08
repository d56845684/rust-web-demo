use actix_web::{web, App, HttpServer, Responder, HttpResponse, HttpRequest};
use actix_files::Files;
use actix_files::NamedFile;
use serde::{Deserialize, Serialize};
// use uuid::Uuid;
use deadpool_postgres::{Pool, Config};
use tokio_postgres::NoTls;
// use deadpool::managed::PoolError;
use log::{info, error};
use jsonwebtoken::{encode, decode, Header, Validation, EncodingKey, DecodingKey};
use chrono::{Utc, Duration};
use once_cell::sync::Lazy;

#[derive(Deserialize, Debug)]
struct UpdateTodo {
    title: Option<String>,
}

#[derive(Deserialize, Debug)]
struct NewTodo {
    title: String,
    done: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
struct Todo {
    id: String,
    title: String,
    done: bool,
    username: String,
}

static JWT_SECRET: Lazy<String> = Lazy::new(|| "secret_key".to_string());

#[derive(Serialize, Deserialize)]
struct Claims {
    sub: String,
    exp: usize,
}

#[derive(Deserialize)]
struct LoginRequest {
    username: String,
    password: String,
}

#[derive(Deserialize)]
struct RegisterRequest {
    username: String,
    password: String,
}

#[derive(Serialize)]
struct LoginResponse {
    token: String,
}

async fn register(info: web::Json<RegisterRequest>, pool: web::Data<Pool>) -> impl Responder {
    match pool.get().await {
        Ok(client) => {
            match client.execute(
                "INSERT INTO users (username, password) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING",
                &[&info.username, &info.password]
            ).await {
                Ok(n) => {
                    if n == 1 {
                        HttpResponse::Ok().finish()
                    } else {
                        HttpResponse::BadRequest().body("User already exists")
                    }
                }
                Err(e) => {
                    error!("Register error: {}", e);
                    HttpResponse::InternalServerError().finish()
                }
            }
        }
        Err(_) => HttpResponse::InternalServerError().finish(),
    }
}

fn authorize(req: &HttpRequest) -> Result<String, HttpResponse> {
    if let Some(header) = req.headers().get("Authorization") {
        if let Ok(auth) = header.to_str() {
            if auth.starts_with("Bearer ") {
                let token = &auth[7..];
                match decode::<Claims>(token, &DecodingKey::from_secret(JWT_SECRET.as_bytes()), &Validation::default()) {
                    Ok(data) => return Ok(data.claims.sub),
                    Err(_) => return Err(HttpResponse::Unauthorized().finish()),
                }
            }
        }
    }
    Err(HttpResponse::Unauthorized().finish())
}

async fn init_db(pool: &Pool) -> Result<(), Box<dyn std::error::Error>> {
    let client = pool.get().await.map_err(|e| Box::new(e) as Box<dyn std::error::Error>)?;
    
    // Create extension if it doesn't exist
    client.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"", &[]).await?;
    // Create users table if it doesn't exist
    client.execute(
        "CREATE TABLE IF NOT EXISTS users (
            username TEXT PRIMARY KEY,
            password TEXT NOT NULL
        )",
        &[]
    ).await?;
    // Create todos table if it doesn't exist
    client.execute(
        "CREATE TABLE IF NOT EXISTS todos (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            title TEXT NOT NULL,
            done BOOLEAN NOT NULL DEFAULT false,
            username TEXT NOT NULL REFERENCES users(username)
        )",
        &[]
    ).await?;
    // Insert default user
    client.execute(
        "INSERT INTO users (username, password) VALUES ('admin', 'password') ON CONFLICT (username) DO NOTHING",
        &[]
    ).await?;

    Ok(())
}

async fn get_todos(req: HttpRequest, pool: web::Data<Pool>) -> impl Responder {
    let username = match authorize(&req) {
        Ok(u) => u,
        Err(e) => return e,
    };
    info!("GET /todos - Fetching todos for {}", username);
    match pool.get().await {
        Ok(client) => {
            match client.query(
                "SELECT id::text, title, done, username FROM todos WHERE username = $1",
                &[&username],
            ).await {
                Ok(rows) => {
                    let todos: Vec<Todo> = rows
                        .iter()
                        .map(|row| Todo {
                            id: row.get(0),
                            title: row.get(1),
                            done: row.get(2),
                            username: row.get(3),
                        })
                        .collect();
                    info!("GET /todos - Successfully fetched {} todos", todos.len());
                    HttpResponse::Ok().json(todos)
                }
                Err(e) => {
                    error!("GET /todos - Database query error: {}", e);
                    HttpResponse::InternalServerError()
                        .json(format!("Database query error: {}", e))
                }
            }
        }
        Err(e) => {
            error!("GET /todos - Database connection error: {}", e);
            HttpResponse::InternalServerError().json(format!("Database connection error: {}", e))
        }
    }
}

async fn add_todo(req: HttpRequest, new: web::Json<NewTodo>, pool: web::Data<Pool>) -> impl Responder {
    let username = match authorize(&req) {
        Ok(u) => u,
        Err(e) => return e,
    };
    info!("POST /todos - Adding new todo: {:?} for {}", new, username);
    match pool.get().await {
        Ok(client) => {
            match client.query_one(
                "INSERT INTO todos (title, done, username) VALUES ($1, $2, $3) RETURNING id::text, title, done, username",
                &[&new.title, &new.done, &username]
            ).await {
                Ok(row) => {
                    let todo = Todo {
                        id: row.get(0),
                        title: row.get(1),
                        done: row.get(2),
                        username: row.get(3),
                    };
                    info!("POST /todos - Successfully added todo with id: {}", todo.id);
                    HttpResponse::Ok().json(todo)
                }
                Err(e) => {
                    error!("POST /todos - Database insert error: {}", e);
                    HttpResponse::InternalServerError().json(format!("Database insert error: {}", e))
                }
            }
        }
        Err(e) => {
            error!("POST /todos - Database connection error: {}", e);
            HttpResponse::InternalServerError().json(format!("Database connection error: {}", e))
        }
    }
}

async fn toggle_done(req: HttpRequest, path: web::Path<String>, pool: web::Data<Pool>) -> impl Responder {
    let username = match authorize(&req) {
        Ok(u) => u,
        Err(e) => return e,
    };
    let id = path.into_inner();
    info!("POST /todos/{}/toggle - Toggling todo status for {}", id, username);
    match pool.get().await {
        Ok(client) => {
            match client.query_one(
                "UPDATE todos SET done = NOT done WHERE id::text = $1 AND username = $2 RETURNING id::text, title, done, username",
                &[&id, &username]
            ).await {
                Ok(row) => {
                    let todo = Todo {
                        id: row.get(0),
                        title: row.get(1),
                        done: row.get(2),
                        username: row.get(3),
                    };
                    info!("POST /todos/{}/toggle - Successfully toggled todo status", id);
                    HttpResponse::Ok().json(todo)
                }
                Err(_) => {
                    error!("POST /todos/{}/toggle - Todo not found", id);
                    HttpResponse::NotFound().finish()
                }
            }
        }
        Err(e) => {
            error!("POST /todos/{}/toggle - Database connection error: {}", id, e);
            HttpResponse::InternalServerError().json(format!("Database connection error: {}", e))
        }
    }
}

async fn delete_todo(req: HttpRequest, path: web::Path<String>, pool: web::Data<Pool>) -> impl Responder {
    let username = match authorize(&req) {
        Ok(u) => u,
        Err(e) => return e,
    };
    let id = path.into_inner();
    info!("DELETE /todos/{} - Deleting todo for {}", id, username);
    match pool.get().await {
        Ok(client) => {
            match client.execute(
                "DELETE FROM todos WHERE id::text = $1 AND username = $2",
                &[&id, &username]
            ).await {
                Ok(rows) => {
                    if rows == 1 {
                        info!("DELETE /todos/{} - Successfully deleted todo", id);
                        HttpResponse::NoContent().finish()
                    } else {
                        error!("DELETE /todos/{} - Todo not found", id);
                        HttpResponse::NotFound().finish()
                    }
                }
                Err(e) => {
                    error!("DELETE /todos/{} - Database delete error: {}", id, e);
                    HttpResponse::InternalServerError().json(format!("Database delete error: {}", e))
                }
            }
        }
        Err(e) => {
            error!("DELETE /todos/{} - Database connection error: {}", id, e);
            HttpResponse::InternalServerError().json(format!("Database connection error: {}", e))
        }
    }
}

async fn update_todo(req: HttpRequest, path: web::Path<String>, item: web::Json<UpdateTodo>, pool: web::Data<Pool>) -> impl Responder {
    let username = match authorize(&req) {
        Ok(u) => u,
        Err(e) => return e,
    };
    let id = path.into_inner();
    info!("PUT /todos/{} - Updating todo for {}: {:?}", id, username, item);
    match pool.get().await {
        Ok(client) => {
            if let Some(new_title) = &item.title {
                match client.query_one(
                    "UPDATE todos SET title = $1 WHERE id::text = $2 AND username = $3 RETURNING id::text, title, done, username",
                    &[new_title, &id, &username]
                ).await {
                    Ok(row) => {
                        let todo = Todo {
                            id: row.get(0),
                            title: row.get(1),
                            done: row.get(2),
                            username: row.get(3),
                        };
                        info!("PUT /todos/{} - Successfully updated todo", id);
                        HttpResponse::Ok().json(todo)
                    }
                    Err(_) => {
                        error!("PUT /todos/{} - Todo not found", id);
                        HttpResponse::NotFound().finish()
                    }
                }
            } else {
                error!("PUT /todos/{} - No title provided for update", id);
                HttpResponse::BadRequest().json("No title provided for update")
            }
        }
        Err(e) => {
            error!("PUT /todos/{} - Database connection error: {}", id, e);
            HttpResponse::InternalServerError().json(format!("Database connection error: {}", e))
        }
    }
}

async fn login(info: web::Json<LoginRequest>, pool: web::Data<Pool>) -> impl Responder {
    match pool.get().await {
        Ok(client) => {
            match client.query_one("SELECT password FROM users WHERE username = $1", &[&info.username]).await {
                Ok(row) => {
                    let stored: String = row.get(0);
                    if stored == info.password {
                        let exp = (Utc::now() + Duration::hours(24)).timestamp() as usize;
                        let claims = Claims { sub: info.username.clone(), exp };
                        match encode(&Header::default(), &claims, &EncodingKey::from_secret(JWT_SECRET.as_bytes())) {
                            Ok(token) => HttpResponse::Ok().json(LoginResponse { token }),
                            Err(e) => {
                                error!("JWT encode error: {}", e);
                                HttpResponse::InternalServerError().finish()
                            }
                        }
                    } else {
                        HttpResponse::Unauthorized().finish()
                    }
                }
                Err(_) => HttpResponse::Unauthorized().finish(),
            }
        }
        Err(_) => HttpResponse::InternalServerError().finish(),
    }
}

async fn test_db_connection(pool: web::Data<Pool>) -> impl Responder {
    info!("GET /test-db - Testing database connection");
    match pool.get().await {
        Ok(client) => {
            match client.query_one("SELECT 1", &[]).await {
                Ok(_) => {
                    info!("GET /test-db - Database connection successful");
                    HttpResponse::Ok().json("Database connection successful!")
                }
                Err(e) => {
                    error!("GET /test-db - Database query error: {}", e);
                    HttpResponse::InternalServerError().json(format!("Database query error: {}", e))
                }
            }
        }
        Err(e) => {
            error!("GET /test-db - Database connection error: {}", e);
            HttpResponse::InternalServerError().json(format!("Database connection error: {}", e))
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Initialize logger
    env_logger::init_from_env(env_logger::Env::new().default_filter_or("info"));
    info!("Starting server...");

    // PostgreSQL connection configuration
    let mut cfg = Config::new();
    cfg.host = Some("postgres".to_string());
    cfg.port = Some(5432);
    cfg.dbname = Some("rust_demo".to_string());
    cfg.user = Some("rust_user".to_string());
    cfg.password = Some("rust_password".to_string());

    let pool = cfg.create_pool(None, NoTls).unwrap();
    
    // Initialize database
    if let Err(e) = init_db(&pool).await {
        error!("Failed to initialize database: {}", e);
        return Err(std::io::Error::new(std::io::ErrorKind::Other, e.to_string()));
    }

    info!("Database initialized successfully");
    let pool = web::Data::new(pool);

    info!("Starting HTTP server at http://0.0.0.0:8080");
    HttpServer::new(move || {
        App::new()
            .app_data(pool.clone())
            .route("/todos", web::get().to(get_todos))
            .route("/todos", web::post().to(add_todo))
            .route("/todos/{id}/toggle", web::post().to(toggle_done))
            .route("/api/login", web::post().to(login))
            .route("/api/register", web::post().to(register))
            .route("/login", web::get().to(|| async { NamedFile::open_async("./static/login.html").await }))
            .route("/register", web::get().to(|| async { NamedFile::open_async("./static/register.html").await }))
            .route("/", web::get().to(|| async {
                NamedFile::open_async("./static/index.html").await
            }))
            .route("/todos/{id}", web::delete().to(delete_todo))
            .route("/todos/{id}", web::put().to(update_todo))
            .route("/test-db", web::get().to(test_db_connection))
            .service(Files::new("/", "./static").index_file("index.html"))
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
}
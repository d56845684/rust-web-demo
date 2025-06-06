use actix_web::{web, App, HttpServer, Responder, HttpResponse};
use actix_files::Files;
use actix_files::NamedFile;
use serde::{Deserialize, Serialize};
// use uuid::Uuid;
use deadpool_postgres::{Pool, Config};
use tokio_postgres::NoTls;
// use deadpool::managed::PoolError;
use log::{info, error};

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
}

async fn init_db(pool: &Pool) -> Result<(), Box<dyn std::error::Error>> {
    let client = pool.get().await.map_err(|e| Box::new(e) as Box<dyn std::error::Error>)?;
    
    // Create extension if it doesn't exist
    client.execute("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"", &[]).await?;
    
    // Create todos table if it doesn't exist
    client.execute(
        "CREATE TABLE IF NOT EXISTS todos (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            title TEXT NOT NULL,
            done BOOLEAN NOT NULL DEFAULT false
        )",
        &[]
    ).await?;
    
    Ok(())
}

async fn get_todos(pool: web::Data<Pool>) -> impl Responder {
    info!("GET /todos - Fetching all todos");
    match pool.get().await {
        Ok(client) => {
            match client.query("SELECT id::text, title, done FROM todos", &[]).await {
                Ok(rows) => {
                    let todos: Vec<Todo> = rows.iter().map(|row| Todo {
                        id: row.get(0),
                        title: row.get(1),
                        done: row.get(2),
                    }).collect();
                    info!("GET /todos - Successfully fetched {} todos", todos.len());
                    HttpResponse::Ok().json(todos)
                }
                Err(e) => {
                    error!("GET /todos - Database query error: {}", e);
                    HttpResponse::InternalServerError().json(format!("Database query error: {}", e))
                }
            }
        }
        Err(e) => {
            error!("GET /todos - Database connection error: {}", e);
            HttpResponse::InternalServerError().json(format!("Database connection error: {}", e))
        }
    }
}

async fn add_todo(new: web::Json<NewTodo>, pool: web::Data<Pool>) -> impl Responder {
    info!("POST /todos - Adding new todo: {:?}", new);
    match pool.get().await {
        Ok(client) => {
            match client.query_one(
                "INSERT INTO todos (title, done) VALUES ($1, $2) RETURNING id::text, title, done",
                &[&new.title, &new.done]
            ).await {
                Ok(row) => {
                    let todo = Todo {
                        id: row.get(0),
                        title: row.get(1),
                        done: row.get(2),
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

async fn toggle_done(path: web::Path<String>, pool: web::Data<Pool>) -> impl Responder {
    let id = path.into_inner();
    info!("POST /todos/{}/toggle - Toggling todo status", id);
    match pool.get().await {
        Ok(client) => {
            match client.query_one(
                "UPDATE todos SET done = NOT done WHERE id::text = $1 RETURNING id::text, title, done",
                &[&id]
            ).await {
                Ok(row) => {
                    let todo = Todo {
                        id: row.get(0),
                        title: row.get(1),
                        done: row.get(2),
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

async fn delete_todo(path: web::Path<String>, pool: web::Data<Pool>) -> impl Responder {
    let id = path.into_inner();
    info!("DELETE /todos/{} - Deleting todo", id);
    match pool.get().await {
        Ok(client) => {
            match client.execute("DELETE FROM todos WHERE id::text = $1", &[&id]).await {
                Ok(_) => {
                    info!("DELETE /todos/{} - Successfully deleted todo", id);
                    HttpResponse::NoContent().finish()
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

async fn update_todo(path: web::Path<String>, item: web::Json<UpdateTodo>, pool: web::Data<Pool>) -> impl Responder {
    let id = path.into_inner();
    info!("PUT /todos/{} - Updating todo: {:?}", id, item);
    match pool.get().await {
        Ok(client) => {
            if let Some(new_title) = &item.title {
                match client.query_one(
                    "UPDATE todos SET title = $1 WHERE id::text = $2 RETURNING id::text, title, done",
                    &[new_title, &id]
                ).await {
                    Ok(row) => {
                        let todo = Todo {
                            id: row.get(0),
                            title: row.get(1),
                            done: row.get(2),
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
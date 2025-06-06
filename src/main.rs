use actix_web::{web, App, HttpServer, Responder, HttpResponse};
use actix_files::Files;
use actix_files::NamedFile;
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::sync::Mutex;


#[derive(Deserialize)]
struct UpdateTodo {
    title: Option<String>,
}

#[derive(Deserialize)]
struct NewTodo {
    title: String,
    done: bool,
}

#[derive(Serialize, Deserialize, Clone)]
struct Todo {
    id: String,
    title: String,
    done: bool,
}

type TodoList = Mutex<Vec<Todo>>;

async fn get_todos(data: web::Data<TodoList>) -> impl Responder {
    let todos = data.lock().unwrap();
    HttpResponse::Ok().json(&*todos)
}

async fn add_todo(new: web::Json<NewTodo>, data: web::Data<TodoList>) -> impl Responder {
    let mut todos = data.lock().unwrap();
    let todo = Todo {
        id: Uuid::new_v4().to_string(),
        title: new.title.clone(),
        done: new.done,
    };
    todos.push(todo.clone());
    HttpResponse::Ok().json(todo)
}

async fn toggle_done(path: web::Path<String>, data: web::Data<TodoList>) -> impl Responder {
    let id = path.into_inner();
    let mut todos = data.lock().unwrap();
    if let Some(todo) = todos.iter_mut().find(|t| t.id == id) {
        todo.done = !todo.done;
        return HttpResponse::Ok().json(todo.clone());
    }
    HttpResponse::NotFound().finish()
}
async fn delete_todo(path: web::Path<String>, data: web::Data<TodoList>) -> impl Responder {
    let id = path.into_inner();
    let mut todos = data.lock().unwrap();
    todos.retain(|t| t.id != id);
    HttpResponse::NoContent().finish()
}

async fn update_todo(path: web::Path<String>, item: web::Json<UpdateTodo>, data: web::Data<TodoList>) -> impl Responder {
    let id = path.into_inner();
    let mut todos = data.lock().unwrap();

    if let Some(todo) = todos.iter_mut().find(|t| t.id == id) {
        if let Some(new_title) = &item.title {
            todo.title = new_title.clone();
        }
        return HttpResponse::Ok().json(todo.clone());
    }

    HttpResponse::NotFound().finish()
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let data = web::Data::new(Mutex::new(Vec::<Todo>::new()));

    HttpServer::new(move || {
        App::new()
            .app_data(data.clone())
            .route("/todos", web::get().to(get_todos))
            .route("/todos", web::post().to(add_todo))
            .route("/todos/{id}/toggle", web::post().to(toggle_done))
            // 首頁 HTML 回傳（單頁應用入口）
            .route("/", web::get().to(|| async {
                NamedFile::open_async("./static/index.html").await
            }))
            .route("/todos/{id}", web::delete().to(delete_todo))
            .route("/todos/{id}", web::put().to(update_todo))
            // // 提供 /static 下的所有檔案（HTML/JS/CSS）靜態服務
            .service(Files::new("/", "./static").index_file("index.html"))
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
}
let todos = [];
let currentFilter = 'all';

async function fetchTodos() {
  const res = await fetch("/todos");
  todos = await res.json();
  renderTodos();
}

function setFilter(filter) {
  currentFilter = filter;
  renderTodos();
}

function renderTodos() {
  const list = document.getElementById("list");
  list.innerHTML = "";

  const filtered = todos.filter(todo => {
    if (currentFilter === 'active') return !todo.done;
    if (currentFilter === 'completed') return todo.done;
    return true;
  });

  filtered.forEach(todo => {
    const li = document.createElement("li");
    if (todo.done) li.classList.add("done");

    const title = document.createElement("span");
    title.textContent = todo.title;
    title.className = "editable";
    title.onclick = () => editTodo(todo);

    const controls = document.createElement("div");

    const toggleBtn = document.createElement("button");
    toggleBtn.textContent = todo.done ? "Undo" : "Done";
    toggleBtn.className = todo.done ? "undo" : "done";
    toggleBtn.onclick = () => toggle(todo.id);

    const deleteBtn = document.createElement("button");
    deleteBtn.textContent = "🗑";
    deleteBtn.onclick = () => remove(todo.id);

    controls.appendChild(toggleBtn);
    controls.appendChild(deleteBtn);

    li.appendChild(title);
    li.appendChild(controls);
    list.appendChild(li);
  });
}

async function addTodo() {
  const input = document.getElementById("new-task");
  const title = input.value.trim();
  if (title === "") return;

  await fetch("/todos", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title, done: false })
  });

  input.value = "";
  fetchTodos();
}

async function toggle(id) {
  await fetch(`/todos/${id}/toggle`, { method: "POST" });
  fetchTodos();
}

async function remove(id) {
  await fetch(`/todos/${id}`, { method: "DELETE" });
  fetchTodos();
}

function editTodo(todo) {
  const newTitle = prompt("Edit task title:", todo.title);
  if (newTitle && newTitle.trim() !== "") {
    fetch(`/todos/${todo.id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title: newTitle.trim() })
    }).then(fetchTodos);
  }
}

fetchTodos();
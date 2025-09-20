const rawBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
const API_BASE = rawBase.replace(/\/$/, '');

async function request(path, { token, method = 'GET', body } = {}) {
  const headers = new Headers();
  if (body !== undefined) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || response.statusText);
  }

  if (response.status === 204) {
    return null;
  }

  const contentType = response.headers.get('Content-Type') || '';
  if (contentType.includes('application/json')) {
    return response.json();
  }
  return response.text();
}

export function loginRequest(username, password) {
  return request('/api/login', {
    method: 'POST',
    body: { username, password },
  });
}

export function registerRequest(username, password) {
  return request('/api/register', {
    method: 'POST',
    body: { username, password },
  });
}

export function fetchTodos(token) {
  return request('/todos', { token });
}

export function createTodo(token, todo) {
  return request('/todos', {
    token,
    method: 'POST',
    body: todo,
  });
}

export function toggleTodo(token, id) {
  return request(`/todos/${id}/toggle`, {
    token,
    method: 'POST',
  });
}

export function updateTodo(token, id, title) {
  return request(`/todos/${id}`, {
    token,
    method: 'PUT',
    body: { title },
  });
}

export function deleteTodo(token, id) {
  return request(`/todos/${id}`, {
    token,
    method: 'DELETE',
  });
}

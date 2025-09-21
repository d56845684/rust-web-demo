const { useCallback, useEffect, useMemo, useState } = React;

const FILTERS = {
  all: 'All',
  active: 'Active',
  completed: 'Completed',
};

function TodoItem({ todo, onToggle, onDelete, onEdit }) {
  return (
    <li className={todo.done ? 'done' : ''}>
      <span className="editable" onClick={() => onEdit(todo)}>
        {todo.title}
        {todo.done && <span className="done-tag">Done</span>}
      </span>
      <div className="controls">
        <button
          type="button"
          className={todo.done ? 'undo' : 'done'}
          onClick={() => onToggle(todo.id)}
        >
          {todo.done ? 'Undo' : 'Done'}
        </button>
        <button type="button" onClick={() => onDelete(todo.id)}>🗑</button>
      </div>
    </li>
  );
}

function FilterButton({ filterKey, label, currentFilter, onClick }) {
  const classes = ['filter-btn'];
  if (currentFilter === filterKey) {
    classes.push('active');
  }
  return (
    <button type="button" className={classes.join(' ')} onClick={() => onClick(filterKey)}>
      {label}
    </button>
  );
}

function TodoApp() {
  const [token, setToken] = useState(() => localStorage.getItem('token'));
  const [todos, setTodos] = useState([]);
  const [filter, setFilter] = useState('all');
  const [newTask, setNewTask] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!token) {
      window.location.href = '/login';
    }
  }, [token]);

  const ensureAuthorized = useCallback(async (response) => {
    if (response.status === 401) {
      localStorage.removeItem('token');
      setToken(null);
      throw new Error('Unauthorized');
    }
    if (!response.ok) {
      const message = (await response.text()) || 'Request failed';
      throw new Error(message);
    }
    return response;
  }, []);

  const fetchTodos = useCallback(async () => {
    if (!token) {
      return;
    }
    setLoading(true);
    setError('');
    try {
      const response = await fetch('/todos', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      const data = await ensureAuthorized(response).then((res) => res.json());
      setTodos(data);
    } catch (err) {
      console.error('Failed to fetch todos', err);
      if (err.message !== 'Unauthorized') {
        setError('Unable to load todos. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  }, [ensureAuthorized, token]);

  useEffect(() => {
    fetchTodos();
  }, [fetchTodos]);

  const filteredTodos = useMemo(() => {
    switch (filter) {
      case 'active':
        return todos.filter((todo) => !todo.done);
      case 'completed':
        return todos.filter((todo) => todo.done);
      default:
        return todos;
    }
  }, [filter, todos]);

  const handleAdd = async (event) => {
    event.preventDefault();
    const title = newTask.trim();
    if (!title || !token) {
      return;
    }
    try {
      setError('');
      const response = await fetch('/todos', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ title, done: false }),
      });
      await ensureAuthorized(response);
      setNewTask('');
      await fetchTodos();
    } catch (err) {
      console.error('Failed to add todo', err);
      if (err.message !== 'Unauthorized') {
        setError('Unable to add the task. Please try again.');
      }
    }
  };

  const handleToggle = async (id) => {
    if (!token) {
      return;
    }
    try {
      setError('');
      const response = await fetch(`/todos/${id}/toggle`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      await ensureAuthorized(response);
      await fetchTodos();
    } catch (err) {
      console.error('Failed to toggle todo', err);
      if (err.message !== 'Unauthorized') {
        setError('Unable to update the task. Please try again.');
      }
    }
  };

  const handleDelete = async (id) => {
    if (!token) {
      return;
    }
    try {
      setError('');
      const response = await fetch(`/todos/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      await ensureAuthorized(response);
      await fetchTodos();
    } catch (err) {
      console.error('Failed to delete todo', err);
      if (err.message !== 'Unauthorized') {
        setError('Unable to delete the task. Please try again.');
      }
    }
  };

  const handleEdit = async (todo) => {
    if (!token) {
      return;
    }
    const updatedTitle = prompt('Edit task title:', todo.title);
    if (!updatedTitle || !updatedTitle.trim()) {
      return;
    }
    try {
      setError('');
      const response = await fetch(`/todos/${todo.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ title: updatedTitle.trim() }),
      });
      await ensureAuthorized(response);
      await fetchTodos();
    } catch (err) {
      console.error('Failed to update todo', err);
      if (err.message !== 'Unauthorized') {
        setError('Unable to update the task. Please try again.');
      }
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setToken(null);
  };

  return (
    <div>
      <h1>Rust To-Do List</h1>
      <form onSubmit={handleAdd} className="input-section">
        <input
          value={newTask}
          placeholder="Add a new task..."
          onChange={(event) => {
            setNewTask(event.target.value);
            if (error) {
              setError('');
            }
          }}
        />
        <button className="add" type="submit" disabled={!newTask.trim()}>
          Add
        </button>
      </form>

      <div className="filters">
        {Object.entries(FILTERS).map(([key, label]) => (
          <FilterButton
            key={key}
            filterKey={key}
            label={label}
            currentFilter={filter}
            onClick={setFilter}
          />
        ))}
      </div>

      <div className="toolbar">
        <button type="button" onClick={handleLogout} className="logout-button">
          Log out
        </button>
      </div>

      {error && <p className="error-message">{error}</p>}
      {loading ? (
        <p className="status-message">Loading...</p>
      ) : filteredTodos.length === 0 ? (
        <p className="status-message">No tasks to display.</p>
      ) : (
        <ul>
          {filteredTodos.map((todo) => (
            <TodoItem
              key={todo.id}
              todo={todo}
              onToggle={handleToggle}
              onDelete={handleDelete}
              onEdit={handleEdit}
            />
          ))}
        </ul>
      )}
    </div>
  );
}

const root = document.getElementById('root');
if (root) {
  const rootInstance = ReactDOM.createRoot(root);
  rootInstance.render(<TodoApp />);
}

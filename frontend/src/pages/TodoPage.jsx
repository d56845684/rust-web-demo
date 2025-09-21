import { useEffect, useMemo, useState } from 'react';
import {
  createTodo,
  deleteTodo,
  fetchTodos,
  toggleTodo,
  updateTodo,
} from '../api.js';
import { useAuth } from '../state/AuthContext.jsx';

const FILTERS = {
  all: () => true,
  active: (todo) => !todo.done,
  completed: (todo) => todo.done,
};

export default function TodoPage() {
  const { token, username, logout } = useAuth();
  const [todos, setTodos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('all');
  const [newTodoTitle, setNewTodoTitle] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let active = true;
    setLoading(true);
    fetchTodos(token)
      .then((items) => {
        if (active) {
          setTodos(items);
          setError('');
        }
      })
      .catch(() => {
        if (active) {
          setError('無法載入待辦事項，請稍候再試');
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [token]);

  const visibleTodos = useMemo(
    () => todos.filter(FILTERS[filter] ?? FILTERS.all),
    [todos, filter],
  );

  const handleCreate = async (event) => {
    event.preventDefault();
    if (!newTodoTitle.trim()) {
      return;
    }
    setSaving(true);
    try {
      const created = await createTodo(token, { title: newTodoTitle.trim(), done: false });
      setTodos((items) => [created, ...items]);
      setNewTodoTitle('');
    } catch (err) {
      setError('新增待辦事項時發生錯誤');
    } finally {
      setSaving(false);
    }
  };

  const handleToggle = async (todo) => {
    try {
      const updated = await toggleTodo(token, todo.id);
      setTodos((items) => items.map((item) => (item.id === updated.id ? updated : item)));
    } catch (err) {
      setError('更新狀態失敗，請重試');
    }
  };

  const handleDelete = async (todo) => {
    try {
      await deleteTodo(token, todo.id);
      setTodos((items) => items.filter((item) => item.id !== todo.id));
    } catch (err) {
      setError('刪除失敗，請稍後再試');
    }
  };

  const handleTitleChange = async (todo, nextTitle) => {
    const trimmed = nextTitle.trim();
    if (!trimmed || trimmed === todo.title) {
      return;
    }
    try {
      const updated = await updateTodo(token, todo.id, trimmed);
      setTodos((items) => items.map((item) => (item.id === updated.id ? updated : item)));
    } catch (err) {
      setError('更新標題時發生錯誤');
    }
  };

  return (
    <div className="app-shell">
      <div className="card">
        <div className="todo-header">
          <div>
            <h1>待辦清單</h1>
            <p>你好，{username || '使用者'}</p>
          </div>
          <div className="todo-actions">
            <div className="filter-group">
              {Object.keys(FILTERS).map((key) => (
                <button
                  key={key}
                  type="button"
                  className={filter === key ? 'active' : ''}
                  onClick={() => setFilter(key)}
                >
                  {key === 'all' ? '全部' : key === 'active' ? '未完成' : '已完成'}
                </button>
              ))}
            </div>
            <button className="secondary" type="button" onClick={logout}>
              登出
            </button>
          </div>
        </div>

        <form onSubmit={handleCreate}>
          <label>
            新增待辦事項
            <input
              type="text"
              value={newTodoTitle}
              onChange={(event) => setNewTodoTitle(event.target.value)}
              placeholder="輸入待辦事項標題"
            />
          </label>
          <button className="primary" type="submit" disabled={saving}>
            {saving ? '建立中…' : '新增'}
          </button>
        </form>

        {error ? <div className="error-text" style={{ marginTop: '1rem' }}>{error}</div> : null}
        {loading ? <p>載入中…</p> : null}

        {!loading && visibleTodos.length === 0 ? <p>目前沒有待辦事項</p> : null}

        <ul className="todo-list">
          {visibleTodos.map((todo) => (
            <li key={todo.id} className={`todo-item${todo.done ? ' completed' : ''}`}>
              <button
                type="button"
                className="secondary"
                onClick={() => handleToggle(todo)}
              >
                {todo.done ? '復原' : '完成'}
              </button>
              <input
                type="text"
                defaultValue={todo.title}
                onBlur={(event) => handleTitleChange(todo, event.target.value)}
              />
              <button type="button" className="secondary" onClick={() => handleDelete(todo)}>
                刪除
              </button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

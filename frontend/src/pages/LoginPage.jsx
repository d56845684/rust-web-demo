import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { loginRequest } from '../api.js';
import { useAuth } from '../state/AuthContext.jsx';

export default function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setLoading(true);
    try {
      const result = await loginRequest(username.trim(), password);
      login(result.token, username.trim());
      navigate('/todos', { replace: true });
    } catch (err) {
      setError('登入失敗，請確認帳號密碼是否正確');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-shell">
      <div className="card">
        <h1>登入</h1>
        <form onSubmit={handleSubmit}>
          <label>
            帳號
            <input
              type="text"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              required
            />
          </label>
          <label>
            密碼
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              required
            />
          </label>
          {error ? <div className="error-text">{error}</div> : null}
          <button className="primary" type="submit" disabled={loading}>
            {loading ? '登入中…' : '登入'}
          </button>
        </form>
        <p>
          還沒有帳號？<Link to="/register">註冊一個</Link>
        </p>
      </div>
    </div>
  );
}

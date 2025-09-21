import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { registerRequest } from '../api.js';

export default function RegisterPage() {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setSuccess('');

    if (password !== confirm) {
      setError('兩次輸入的密碼不一致');
      return;
    }

    setLoading(true);
    try {
      await registerRequest(username.trim(), password);
      setSuccess('註冊成功，請使用新帳號登入');
      setTimeout(() => navigate('/login'), 800);
    } catch (err) {
      setError('註冊失敗，請稍後再試或更換帳號');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-shell">
      <div className="card">
        <h1>註冊</h1>
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
              autoComplete="new-password"
              required
            />
          </label>
          <label>
            確認密碼
            <input
              type="password"
              value={confirm}
              onChange={(event) => setConfirm(event.target.value)}
              autoComplete="new-password"
              required
            />
          </label>
          {error ? <div className="error-text">{error}</div> : null}
          {success ? <div className="success-text">{success}</div> : null}
          <button className="primary" type="submit" disabled={loading}>
            {loading ? '註冊中…' : '註冊'}
          </button>
        </form>
        <p>
          已經有帳號了嗎？<Link to="/login">立即登入</Link>
        </p>
      </div>
    </div>
  );
}

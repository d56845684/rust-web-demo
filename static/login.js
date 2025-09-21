const { useEffect, useState } = React;

function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (localStorage.getItem('token')) {
      window.location.href = '/';
    }
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setLoading(true);
    setError('');
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      if (!response.ok) {
        const message = (await response.text()) || 'Login failed. Please check your credentials.';
        throw new Error(message);
      }
      const data = await response.json();
      localStorage.setItem('token', data.token);
      window.location.href = '/';
    } catch (err) {
      console.error('Login failed', err);
      setError(err.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>Login</h1>
      <form className="input-section" onSubmit={handleSubmit}>
        <input
          value={username}
          placeholder="Username"
          autoComplete="username"
          onChange={(event) => {
            setUsername(event.target.value);
            if (error) setError('');
          }}
        />
        <input
          type="password"
          value={password}
          placeholder="Password"
          autoComplete="current-password"
          onChange={(event) => {
            setPassword(event.target.value);
            if (error) setError('');
          }}
        />
        <button className="add" type="submit" disabled={loading || !username || !password}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>
      {error && <p className="error-message">{error}</p>}
      <div className="auth-links">
        <span>
          Don't have an account? <a href="/register">Register</a>
        </span>
      </div>
    </div>
  );
}

const root = document.getElementById('root');
if (root) {
  ReactDOM.createRoot(root).render(<LoginPage />);
}


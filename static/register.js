const { useEffect, useState } = React;

function RegisterPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [status, setStatus] = useState('');
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
    setStatus('');
    try {
      const response = await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      if (!response.ok) {
        const message = (await response.text()) || 'Registration failed. Please try again.';
        throw new Error(message);
      }
      setStatus('Registration successful! Redirecting to login...');
      setTimeout(() => {
        window.location.href = '/login';
      }, 1200);
    } catch (err) {
      console.error('Registration failed', err);
      setError(err.message || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>Register</h1>
      <form className="input-section" onSubmit={handleSubmit}>
        <input
          value={username}
          placeholder="Username"
          autoComplete="username"
          onChange={(event) => {
            setUsername(event.target.value);
            if (error) setError('');
            if (status) setStatus('');
          }}
        />
        <input
          type="password"
          value={password}
          placeholder="Password"
          autoComplete="new-password"
          onChange={(event) => {
            setPassword(event.target.value);
            if (error) setError('');
            if (status) setStatus('');
          }}
        />
        <button className="add" type="submit" disabled={loading || !username || !password}>
          {loading ? 'Registering...' : 'Register'}
        </button>
      </form>
      {error && <p className="error-message">{error}</p>}
      {status && <p className="status-message">{status}</p>}
      <div className="auth-links">
        <span>
          Already have an account? <a href="/login">Login</a>
        </span>
      </div>
    </div>
  );
}

const root = document.getElementById('root');
if (root) {
  ReactDOM.createRoot(root).render(<RegisterPage />);
}


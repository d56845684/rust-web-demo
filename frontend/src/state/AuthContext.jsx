import { createContext, useContext, useMemo, useState } from 'react';

const AuthContext = createContext(null);
const TOKEN_KEY = 'rust-web-demo-token';
const USERNAME_KEY = 'rust-web-demo-username';

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) ?? '');
  const [username, setUsername] = useState(() => localStorage.getItem(USERNAME_KEY) ?? '');

  const login = (nextToken, nextUsername) => {
    localStorage.setItem(TOKEN_KEY, nextToken);
    localStorage.setItem(USERNAME_KEY, nextUsername);
    setToken(nextToken);
    setUsername(nextUsername);
  };

  const logout = () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USERNAME_KEY);
    setToken('');
    setUsername('');
  };

  const value = useMemo(() => ({
    token,
    username,
    isAuthenticated: Boolean(token),
    login,
    logout,
  }), [token, username]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used inside AuthProvider');
  }
  return ctx;
}

import { Navigate, Route, Routes } from 'react-router-dom';
import LoginPage from './pages/LoginPage.jsx';
import RegisterPage from './pages/RegisterPage.jsx';
import TodoPage from './pages/TodoPage.jsx';
import { useAuth } from './state/AuthContext.jsx';

function PrivateRoute({ children }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return children;
}

export default function App() {
  const { isAuthenticated } = useAuth();

  return (
    <Routes>
      <Route
        path="/login"
        element={isAuthenticated ? <Navigate to="/todos" replace /> : <LoginPage />}
      />
      <Route
        path="/register"
        element={isAuthenticated ? <Navigate to="/todos" replace /> : <RegisterPage />}
      />
      <Route
        path="/"
        element={<Navigate to={isAuthenticated ? '/todos' : '/login'} replace />}
      />
      <Route
        path="/todos"
        element={(
          <PrivateRoute>
            <TodoPage />
          </PrivateRoute>
        )}
      />
      <Route
        path="*"
        element={<Navigate to={isAuthenticated ? '/todos' : '/login'} replace />}
      />
    </Routes>
  );
}

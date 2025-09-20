import { Navigate, Outlet, Route, Routes } from 'react-router-dom';
import LoginPage from './pages/LoginPage.jsx';
import RegisterPage from './pages/RegisterPage.jsx';
import TodoPage from './pages/TodoPage.jsx';
import { useAuth } from './state/AuthContext.jsx';

function PrivateRoute() {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
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
      <Route element={<PrivateRoute />}>
        <Route path="/todos" element={<TodoPage />} />
      </Route>
      <Route
        path="*"
        element={<Navigate to={isAuthenticated ? '/todos' : '/login'} replace />}
      />
    </Routes>
  );
}

import { Navigate } from 'react-router-dom';
import { Layout } from './Layout';
import { isAuthenticated } from './auth-utils';

/**
 * Props for ProtectedRoute component
 */
interface ProtectedRouteProps {
  children: React.ReactNode;
}

/**
 * ProtectedRoute component
 * Redirects to login if user is not authenticated
 * Wraps authenticated content with Layout (includes Header)
 */
export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }

  return <Layout>{children}</Layout>;
};

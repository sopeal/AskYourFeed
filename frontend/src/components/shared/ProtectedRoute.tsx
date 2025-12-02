import { Navigate } from 'react-router-dom';
import { ReactNode } from 'react';

/**
 * Props for ProtectedRoute component
 */
interface ProtectedRouteProps {
  children: ReactNode;
}

/**
 * Check if user is authenticated
 * Checks for session token in localStorage
 */
const isAuthenticated = (): boolean => {
  const token = localStorage.getItem('session_token');
  const user = localStorage.getItem('user');
  
  if (!token || !user) {
    return false;
  }

  try {
    const userData = JSON.parse(user);
    const expiresAt = new Date(userData.session_expires_at);
    const now = new Date();
    
    // Check if session has expired
    if (expiresAt <= now) {
      // Clear expired session
      localStorage.removeItem('session_token');
      localStorage.removeItem('user');
      return false;
    }
    
    return true;
  } catch (error) {
    // Invalid user data, clear storage
    localStorage.removeItem('session_token');
    localStorage.removeItem('user');
    return false;
  }
};

/**
 * ProtectedRoute component
 * Redirects to login if user is not authenticated
 */
export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

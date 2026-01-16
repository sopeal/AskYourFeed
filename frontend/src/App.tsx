import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RegisterView } from './views/RegisterView';
import { LoginView } from './views/LoginView';
import { DashboardView } from './views/DashboardView';
import { HistoryView } from './views/HistoryView';
import { ProtectedRoute } from './components/shared/ProtectedRoute';
import { isAuthenticated } from './components/shared/auth-utils';
import './App.css';

// Create a client for React Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

/**
 * AuthRoute component
 * Redirects to dashboard if user is already authenticated
 */
const AuthRoute = ({ children }: { children: React.ReactNode }) => {
  if (isAuthenticated()) {
    return <Navigate to="/" replace />;
  }
  return <>{children}</>;
};

/**
 * Main App component with routing and providers
 */
function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          {/* Authentication routes - redirect to dashboard if already logged in */}
          <Route 
            path="/register" 
            element={
              <AuthRoute>
                <RegisterView />
              </AuthRoute>
            } 
          />
          <Route 
            path="/login" 
            element={
              <AuthRoute>
                <LoginView />
              </AuthRoute>
            } 
          />
          
          {/* Protected dashboard route */}
          <Route 
            path="/" 
            element={
              <ProtectedRoute>
                <DashboardView />
              </ProtectedRoute>
            } 
          />
          
          {/* Protected history route */}
          <Route 
            path="/history" 
            element={
              <ProtectedRoute>
                <HistoryView />
              </ProtectedRoute>
            } 
          />
          
          {/* Catch all - redirect to home */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;

import { LoginForm } from './components/LoginForm';

/**
 * Login page view
 * Main container for the login form
 */
export const LoginView = () => {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4 py-12">
      <div className="w-full max-w-md space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
            Zaloguj się
          </h1>
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            Witaj ponownie! Zaloguj się do swojego konta Ask Your Feed
          </p>
        </div>

        {/* Login Form */}
        <div className="bg-white dark:bg-gray-800 shadow-md rounded-lg px-8 py-10">
          <LoginForm />
        </div>
      </div>
    </div>
  );
};

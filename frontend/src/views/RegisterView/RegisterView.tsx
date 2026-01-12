import { RegisterForm } from './components/RegisterForm';

/**
 * Registration page view
 * Main container for the registration form
 */
export const RegisterView = () => {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4 py-12">
      <div className="w-full max-w-md space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
            Stwórz konto
          </h1>
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            Dołącz do Ask Your Feed i zacznij zadawać pytania swojemu feedowi X
          </p>
        </div>

        {/* Registration Form */}
        <div className="bg-white dark:bg-gray-800 shadow-md rounded-lg px-8 py-10">
          <RegisterForm />
        </div>
      </div>
    </div>
  );
};

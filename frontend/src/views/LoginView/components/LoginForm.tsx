import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { loginSchema, type LoginFormData } from '../../../schemas/auth.schema';
import { useLogin } from '../../../hooks/useLogin';
import { InputField } from '../../../components/shared/InputField';
import { SubmitButton } from '../../../components/shared/SubmitButton';
import type { LoginCommand } from '../../../types/auth.types';
import { AxiosError } from 'axios';
import type { ErrorResponseDTO } from '../../../types/auth.types';

/**
 * Login form component
 * Handles user login with validation and API integration
 */
export const LoginForm = () => {
  const navigate = useNavigate();
  const { mutate: login, isPending } = useLogin();

  const {
    register: registerField,
    handleSubmit,
    formState: { errors },
    setError,
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  /**
   * Handle form submission
   */
  const onSubmit = (data: LoginFormData) => {
    // Map form data to API request format
    const loginCommand: LoginCommand = {
      email: data.email,
      password: data.password,
    };

    login(loginCommand, {
      onSuccess: (response) => {
        // TODO: Save session token and user data to AuthContext
        // For now, we'll store in localStorage
        localStorage.setItem('session_token', response.session_token);
        localStorage.setItem('user', JSON.stringify({
          user_id: response.user_id,
          email: response.email,
          x_username: response.x_username,
          x_display_name: response.x_display_name,
          session_expires_at: response.session_expires_at,
        }));

        // Redirect to main dashboard
        navigate('/');
      },
      onError: (error: AxiosError<ErrorResponseDTO>) => {
        // Handle API errors
        if (error.response?.status === 401) {
          // Unauthorized - invalid credentials
          setError('root', {
            type: 'manual',
            message: 'Nieprawidłowy adres e-mail lub hasło.',
          });
        } else if (error.response?.status === 500 || error.response?.status === 503) {
          // Server error
          setError('root', {
            type: 'manual',
            message: 'Wystąpił błąd serwera. Spróbuj ponownie później.',
          });
        } else if (error.response?.data?.error) {
          // Other API errors with error details
          const errorMessage = error.response.data.error.message;
          setError('root', {
            type: 'manual',
            message: errorMessage || 'Wystąpił błąd podczas logowania. Spróbuj ponownie.',
          });
        } else {
          // Generic error
          setError('root', {
            type: 'manual',
            message: 'Wystąpił nieoczekiwany błąd. Spróbuj ponownie.',
          });
        }
      },
    });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Root error message */}
      {errors.root && (
        <div className="p-3 text-sm text-red-500 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900 rounded-md" role="alert">
          {errors.root.message}
        </div>
      )}

      {/* Email field */}
      <InputField
        id="email"
        label="Adres e-mail"
        type="email"
        autoComplete="email"
        error={errors.email?.message}
        {...registerField('email')}
      />

      {/* Password field */}
      <InputField
        id="password"
        label="Hasło"
        type="password"
        autoComplete="current-password"
        error={errors.password?.message}
        {...registerField('password')}
      />

      {/* Submit button */}
      <SubmitButton
        isLoading={isPending}
        loadingText="Logowanie..."
        className="w-full"
      >
        Zaloguj się
      </SubmitButton>

      {/* Link to register */}
      <p className="text-center text-sm text-gray-600 dark:text-gray-400">
        Nie masz jeszcze konta?{' '}
        <a
          href="/register"
          className="font-medium text-primary hover:underline"
        >
          Zarejestruj się
        </a>
      </p>
    </form>
  );
};

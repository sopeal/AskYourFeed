import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { registerSchema, type RegisterFormData } from '@/schemas/auth.schema.ts';
import { useRegister } from '@/hooks/useRegister.ts';
import { InputField } from '@/components/shared/InputField.tsx';
import { SubmitButton } from '@/components/shared/SubmitButton.tsx';
import type { RegisterCommand } from '@/types/auth.types.ts';
import { AxiosError } from 'axios';
import type { ErrorResponseDTO } from '@/types/auth.types.ts';

/**
 * Registration form component
 * Handles user registration with validation and API integration
 */
export const RegisterForm = () => {
  const navigate = useNavigate();
  const { mutate: register, isPending } = useRegister();

  const {
    register: registerField,
    handleSubmit,
    formState: { errors },
    setError,
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      password: '',
      passwordConfirmation: '',
      xUsername: '',
    },
  });

  /**
   * Handle form submission
   */
  const onSubmit = (data: RegisterFormData) => {
    // Map form data to API request format
    const registerCommand: RegisterCommand = {
      email: data.email,
      password: data.password,
      password_confirmation: data.passwordConfirmation,
      x_username: data.xUsername,
    };

    register(registerCommand, {
      onSuccess: (response) => {
        // TODO: Save session token and user data to AuthContext
        // For now, we'll store in localStorage
        localStorage.setItem('session_token', response.session_token);
        localStorage.setItem('user', JSON.stringify({
          user_id: response.user_id,
          email: response.email,
          x_username: response.x_username,
          x_display_name: response.x_display_name,
        }));

        // Redirect to main dashboard
        navigate('/');
      },
      onError: (error: AxiosError<ErrorResponseDTO>) => {
        // Handle API errors
        if (error.response?.data?.error) {
          const errorCode = error.response.data.error.code;
          const errorMessage = error.response.data.error.message;

          switch (errorCode) {
            case 'EMAIL_ALREADY_REGISTERED':
              setError('email', {
                type: 'manual',
                message: 'Podany adres e-mail jest już zajęty.',
              });
              break;
            case 'X_USERNAME_NOT_FOUND':
              setError('xUsername', {
                type: 'manual',
                message: `Konto X o nazwie '${registerCommand.x_username}' nie istnieje. Sprawdź poprawność nazwy użytkownika.`,
              });
              break;
            case 'VALIDATION_ERROR':
              // Generic validation error
              setError('root', {
                type: 'manual',
                message: errorMessage || 'Nieprawidłowe dane.',
              });
              break;
            default:
              setError('root', {
                type: 'manual',
                message: errorMessage || 'Wystąpił błąd podczas rejestracji. Spróbuj ponownie.',
              });
          }
        } else if (error.response?.status === 500 || error.response?.status === 503) {
          setError('root', {
            type: 'manual',
            message: 'Wystąpił błąd serwera. Spróbuj ponownie później.',
          });
        } else {
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
        autoComplete="new-password"
        error={errors.password?.message}
        {...registerField('password')}
      />

      {/* Password confirmation field */}
      <InputField
        id="passwordConfirmation"
        label="Potwierdź hasło"
        type="password"
        autoComplete="new-password"
        error={errors.passwordConfirmation?.message}
        {...registerField('passwordConfirmation')}
      />

      {/* X Username field */}
      <InputField
        id="xUsername"
        label="Nazwa użytkownika X (Twitter)"
        type="text"
        placeholder="np. username"
        error={errors.xUsername?.message}
        {...registerField('xUsername')}
      />

      {/* Submit button */}
      <SubmitButton
        isLoading={isPending}
        loadingText="Rejestrowanie..."
        className="w-full"
      >
        Zarejestruj się
      </SubmitButton>

      {/* Link to login */}
      <p className="text-center text-sm text-gray-600 dark:text-gray-400">
        Masz już konto?{' '}
        <a
          href="/login"
          className="font-medium text-primary hover:underline"
        >
          Zaloguj się
        </a>
      </p>
    </form>
  );
};

import { z } from 'zod';

/**
 * Registration form validation schema
 * Validates client-side before sending to API
 */
export const registerSchema = z.object({
  email: z
    .string()
    .min(1, 'Adres e-mail jest wymagany')
    .email('Podaj poprawny adres e-mail'),
  
  password: z
    .string()
    .min(8, 'Hasło musi mieć co najmniej 8 znaków')
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^A-Za-z0-9])\S+$/,
      'Hasło musi zawierać co najmniej jedną wielką literę, jedną małą literę, jedną cyfrę i jeden znak specjalny'
    ),
  
  passwordConfirmation: z
    .string()
    .min(1, 'Potwierdzenie hasła jest wymagane'),
  
  xUsername: z
    .string()
    .min(1, 'Nazwa użytkownika X jest wymagana')
    .regex(/^[A-Za-z0-9_]+$/, 'Nazwa użytkownika X może zawierać tylko litery, cyfry i podkreślenia')
}).refine((data) => data.password === data.passwordConfirmation, {
  message: 'Hasła nie są zgodne',
  path: ['passwordConfirmation'],
});

/**
 * Login form validation schema
 */
export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'Adres e-mail jest wymagany')
    .email('Podaj poprawny adres e-mail'),
  
  password: z
    .string()
    .min(1, 'Hasło jest wymagane'),
});

export type RegisterFormData = z.infer<typeof registerSchema>;
export type LoginFormData = z.infer<typeof loginSchema>;

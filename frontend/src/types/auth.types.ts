// =============================================================================
// Authentication DTOs - matching backend/internal/dto/dto.go
// =============================================================================

/**
 * POST /api/v1/auth/register - Request Body
 */
export interface RegisterCommand {
  email: string;
  password: string;
  password_confirmation: string;
  x_username: string;
}

/**
 * POST /api/v1/auth/register - Success Response
 */
export interface RegisterResponseDTO {
  user_id: string; // UUID
  email: string;
  x_username: string;
  x_display_name: string;
  created_at: string; // ISO 8601 Date
  session_token: string;
}

/**
 * POST /api/v1/auth/login - Request Body
 */
export interface LoginCommand {
  email: string;
  password: string;
}

/**
 * POST /api/v1/auth/login - Success Response
 */
export interface LoginResponseDTO {
  user_id: string;
  email: string;
  x_username: string;
  x_display_name: string;
  session_token: string;
  session_expires_at: string; // ISO 8601 Date
}

// =============================================================================
// Error Response DTOs
// =============================================================================

/**
 * Standard API Error Response
 */
export interface ErrorResponseDTO {
  error: ErrorDetailDTO;
}

/**
 * Error details
 */
export interface ErrorDetailDTO {
  code: string;
  message: string;
  details?: Record<string, any>;
}

// =============================================================================
// View Models - Internal component state
// =============================================================================

/**
 * Registration form state
 */
export interface RegisterFormViewModel {
  email: string;
  password: string;
  passwordConfirmation: string;
  xUsername: string;
}

/**
 * Registration form validation errors
 */
export type RegisterFormValidation = {
  [key in keyof RegisterFormViewModel]?: string;
};

/**
 * Login form state
 */
export interface LoginFormViewModel {
  email: string;
  password: string;
}

/**
 * Login form validation errors
 */
export type LoginFormValidation = {
  [key in keyof LoginFormViewModel]?: string;
};

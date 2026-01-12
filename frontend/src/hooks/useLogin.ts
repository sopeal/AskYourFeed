import { useMutation } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';
import type { LoginCommand, LoginResponseDTO, ErrorResponseDTO } from '../types/auth.types';

/**
 * API client configuration
 */
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

/**
 * Login user
 * POST /api/v1/auth/login
 */
const loginUser = async (data: LoginCommand): Promise<LoginResponseDTO> => {
  const response = await axios.post<LoginResponseDTO>(
    `${API_BASE_URL}/api/v1/auth/login`,
    data,
    {
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );
  return response.data;
};

/**
 * Hook for user login
 * Handles API call, loading state, and error handling
 */
export const useLogin = () => {
  return useMutation<LoginResponseDTO, AxiosError<ErrorResponseDTO>, LoginCommand>({
    mutationFn: loginUser,
  });
};

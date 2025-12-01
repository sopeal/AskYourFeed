import { useMutation } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';
import type { RegisterCommand, RegisterResponseDTO, ErrorResponseDTO } from '../types/auth.types';

/**
 * API client configuration
 */
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

/**
 * Register a new user
 * POST /api/v1/auth/register
 */
const registerUser = async (data: RegisterCommand): Promise<RegisterResponseDTO> => {
  const response = await axios.post<RegisterResponseDTO>(
    `${API_BASE_URL}/api/v1/auth/register`,
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
 * Hook for user registration
 * Handles API call, loading state, and error handling
 */
export const useRegister = () => {
  return useMutation<RegisterResponseDTO, AxiosError<ErrorResponseDTO>, RegisterCommand>({
    mutationFn: registerUser,
  });
};
